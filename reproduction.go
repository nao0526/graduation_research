package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
	"encoding/csv"
	"log"
	"os"
	"strconv"
)

type Player struct {
	endowment float64
	fitness float64
	payoff float64
	strategies []map[string]float64
	count int
}

type Players []Player

func (player *Player) init(E float64, R int, T float64) {
	player.endowment = E
	player.fitness = 0
	player.payoff = 0
	player.count = 0
	for r := 0; r < R; r++ {
		var tau float64
		var j float64
		var k float64
		tau = rand.Float64() * T
		j = 0
		if rand.Intn(2) == 1 {
			j = E / float64(R)
		}
		k = 0
		if rand.Intn(2) == 1 {
			k = E / float64(R)
		}
		player.strategies = append(player.strategies, map[string]float64{"tau": tau, "j": j, "k": k})
	}
}

func (player *Player) initNext(E float64) {
	player.updateEndowment(E)
	player.updateFitness(0)
	player.updatePayoff(0)
	player.updateCount(0)
}

func (player *Player) updateEndowment(newEndowment float64) {
	player.endowment = newEndowment
}

func (player *Player) calcFitness() float64 {
	var beta float64 = 1.0
	return math.Exp(beta * player.payoff)
}

func (player *Player) updateFitness(fitness float64) {
	player.fitness = fitness
}

func (player *Player) updatePayoff(payoff float64) {
	player.payoff = payoff
}

func (player *Player) updateCount(count int) {
	player.count = count
}

func (player *Player) updateStrategies(strategies []map[string]float64) {
	player.strategies = strategies
}

func allKeys(m map[int]bool) []int {
	result := make([]int, len(m))
	i := 0
	for key, _ := range m {
		result[i] = key
		i++
	}
	return result
}

func pickup(min int, max int, num int) []int {
	var numRange int = max - min
	selected := make(map[int]bool)
	for i := 0; i < num; {
		var n int = rand.Intn(numRange) + min
		if selected[n] == false {
			selected[n] = true
			i++
		}
	}
	keys := allKeys(selected)
	sort.Sort(sort.IntSlice(keys))
	return keys;
}

func main() {
	var G, N, M, R, generation, trials int = 10000, 1000, 6, 10, 1000, 100
	var E float64 = 1.0
	risks := [11] float64{0.0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0}
	// risks := [1] float64{1.0}
	var payoffs []float64
	var contributions []float64
	var fines []float64
	var targets []float64
	fileNames := [5] string{"reproduction", "generation", "contribution", "fine", "target"}
	payoff_generations := make([][]float64, len(risks))
	for i :=0; i < len(risks); i++{
		payoff_generations[i] = make([]float64, 201)
	}
	var T float64 = E * float64(M) / 2

	start := time.Now()
	// 乱数のシード値
	rand.Seed(time.Now().UnixNano())
	for q, risk := range risks {
		var sumPayoff float64 = 0.0
		var sumContribution float64 = 0.0
		var sumFine float64 = 0.0
		var sumTarget float64 = 0.0
		// var payoffs_trials [100][201]float64
		for t := 0; t < trials; t++ {
			var players Players // playerを格納しておく配列
			// playerの生成
			for i := 0; i < N; i++ { 
				player := Player{}
				player.init(E, R, T)
				players = append(players, player)
			}
			for g := 0; g < generation; g++ {
				for i := 0; i < G; i++ {
					// ゲームに参加するプレイヤーを決定
					group := pickup(0, N, M)
					// ゲームに参加するプレイヤーの資金の初期化と参加した回数を記録
					for _, p := range group {
						players[p].updateEndowment(E)
						players[p].updateCount(players[p].count + 1)
					}
		
					// ゲーム開始
					var commonPool float64 = 0.0 // 集まったお金をカウントする変数
					var commonPoolNotFine float64 = 0.0 // 集まったお金をカウントする変数
					for r := 0; r < R; r++ {
						var contributionPerRound float64 = 0.0 // 各ラウンドでの寄付額の合計
						for _, p := range group {
							var contribution float64 // プレイヤーの寄付額
							if commonPool >= players[p].strategies[r]["tau"] { // rラウンドまでに集まった寄付額が閾値を超えていた場合
								if players[p].endowment >= players[p].strategies[r]["j"] {
									contribution = players[p].strategies[r]["j"]
								} else {
									contribution = players[p].endowment 
								}
								
							} else { // 超えていなかった場合
								if players[p].endowment >= players[p].strategies[r]["k"] {
									contribution = players[p].strategies[r]["k"]
								} else {
									contribution = players[p].endowment 
								}
							}
							contributionPerRound += contribution 
							players[p].updateEndowment(players[p].endowment - contribution) // 寄付した額だけ資金を減らす
						}
						commonPool += contributionPerRound
						commonPoolNotFine += contributionPerRound
						if r == 4 {
							if commonPool < (T / 2.0) {
								for _, p := range group {
									// var restEndowment float64 = players[p].endowment - (E / 2.0)
									var restEndowment float64 = players[p].endowment
									if restEndowment > 0 {
										var fine float64 = (risk / 2.0) * restEndowment
										players[p].updateEndowment(players[p].endowment - fine)
										commonPool += fine
										if g == generation - 1 {
											sumFine += fine
										}
									}
								}
							}
						}
						// if commonPool < (T / float64(R))  {
						// 	for _, p := range group {
						// 		var restEndowment float64 = players[p].endowment - ((E / float64(R)) * float64(R - (r + 1)))
						// 		if restEndowment > 0 {
						// 			var fine float64 = (risk / float64(R)) * restEndowment
						// 			players[p].updateEndowment(players[p].endowment - fine)
						// 			commonPool += fine
						// 			if g == generation - 1 {
						// 				sumFine += fine
						// 			}
						// 		}
						// 	}
						// }
					}
					if g == generation - 1 {
						sumContribution += commonPoolNotFine
						sumTarget += commonPool
					}

					// 目標額があつまらなかった場合、確率に従い資金損失
					if commonPool < T {
						if rand.Float64() < risk {
							for _, p := range group {
								players[p].updateEndowment(0)
							}
						}
					}
	
					// 余った資金を利得に加算
					for _, p := range group {
						players[p].updatePayoff(players[p].payoff + players[p].endowment)
					}
					
				}
	
				// 平均利得の計算
				var payoff_generation float64 = 0.0
				for i := 0; i < N; i++ {
					players[i].updatePayoff(players[i].payoff / float64(players[i].count))
					payoff_generation += players[i].payoff
					if g == generation - 1 {
						sumPayoff += players[i].payoff
					}
				}
				if g < 201 {
					payoff_generations[q][g] += payoff_generation / float64(N * trials)
				}
				
				
				// 適応度の計算
				for i := 0; i < N; i++ {
					players[i].updateFitness(players[i].calcFitness())
				}
				
	
				// 進化
				var sumFitness float64 = 0.0
				var nextPlayers Players
				for i := 0; i < N; i++ {
					sumFitness += players[i].fitness
				}
				for i := 0; i < N; i++ {
					var random float64 = rand.Float64()
					var rate float64 = 0.0
					for j := 0; j < N; j++ {
						rate += players[j].fitness / sumFitness
						if random < rate {
							// fmt.Println(g, j)
							nextPlayers = append(nextPlayers, players[j])
							break
						}
					} 
				}
				// 戦略を更新
				players = nextPlayers
		
				// 次の世代を初期化
				for i := 0; i < N; i++ {
					players[i].initNext(E)
				}
			}
		}
		payoffs = append(payoffs, sumPayoff / float64(N * trials))
		contributions = append(contributions, sumContribution / float64(M * G * trials))
		fines = append(fines, sumFine / float64(M * G * trials))
		targets = append(targets, sumTarget/ float64(G * trials))

		// for g := 0; g < 201; g++ {
		// 	var payoff_generation float64 = 0.0
		// 	for t := 0; t < trials; t++ {
		// 		payoff_generation += payoffs_trials[t][g]
		// 	}
		// 	payoff_generations[q][g] = payoff_generation / float64(N * trials)
		// }
		
	}	
	fmt.Println(contributions)
	// fmt.Println(payoffs)
	// fmt.Println(payoffs_generation[1])
	end := time.Now()
	fmt.Printf("%f秒\n", (end.Sub(start)).Seconds())

	// ファイル出力
	for _, fileName := range fileNames {
		var records []string
		// // var record []string
		if fileName == "reproduction" {
			for i := 0; i < len(risks); i++ {
				records = append(records, strconv.FormatFloat(payoffs[i], 'f', -1, 64))
			}
		}

		if fileName == "generation" {
			for g:= 0; g < 201; g++ {
				records = append(records, strconv.FormatFloat(payoff_generations[7][g], 'f', -1, 64))
			}
		}

		if fileName == "contribution" {
			for i := 0; i < len(risks); i++ {
				records = append(records, strconv.FormatFloat(contributions[i], 'f', -1, 64))
			}
		}

		if fileName == "fine" {
			for i := 0; i < len(risks); i++ {
				records = append(records, strconv.FormatFloat(fines[i], 'f', -1, 64))
			}
		}

		if fileName == "target" {
			for i := 0; i < len(risks); i++ {
				records = append(records, strconv.FormatFloat(targets[i], 'f', -1, 64))
			}
		}

		f, err := os.Create("./" + fileName + ".csv")
		if err != nil {
			 log.Fatal(err)
		}
	
		w := csv.NewWriter(f)
	
		if err := w.Write(records); err != nil {
			 log.Fatal(err)
		}
	
		w.Flush()
	
		if err := w.Error(); err != nil {
			log.Fatal(err)
		}

	}
}
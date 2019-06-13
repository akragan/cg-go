package main

import "fmt"
import "os"
import "time"

//import "bufio"
//import "strings"

//GLOBAL OPTIONS
const (
	MAKE_LOSING_PLAYER_DISAPPEAR_FROM_SIM_GRID = true
	SCORE_EVAL_SUMOP_DIVISOR                   = 1  // divide sum of opponent points by this divisor
	SCORE_EVAL_NEIGHBOR_DIVISOR                = 10 // divide free neighbor penalty points by this divisor
	PENALIZE_EDGE                              = true
	PENALIZE_EDGE_POINTS                       = 1 // deduct this many points per edge cell
	SCENARIO_DEPTH                             = 6

	// ideas that didn't work...
	RESUME_FLOODFILL_WHEN_PLAYER_DISAPPEARS = false
	SCORE_EVAL_MAXOP                        = false
	PENALIZE_EDGE_WHEN_3_OR_4               = false //BEST: false
	PENALIZE_OP                             = false
	PENALIZE_OP_POINTS                      = 1
)

var (
	DEBUG_SCENARIO = []string{}
	//DEBUG_SCENARIO:=[]string{"DOWN","LEFT","LEFT","LEFT"}

)

// define matrix function constructing and initializing 2-dim array
func new_matrix(m int, n int, initial int) [][]int {
	mat := make([][]int, m)
	for i := 0; i < m; i++ {
		mat[i] = make([]int, n)
		for j := 0; j < n; j++ {
			mat[i][j] = initial
		}
	}
	return mat
}

// copy 2-dim array to another array of same size
func matrix_copy_to(m int, n int, to [][]int, original [][]int) {
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			to[i][j] = original[i][j]
		}
	}
}

// overwrite
func matrix_overwrite(m int, n int, mat [][]int, orig_value int, new_value int) {
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			if mat[i][j] == orig_value {
				mat[i][j] = new_value
			}
		}
	}
}

// define Position object prototype
type Pos struct {
	X int
	Y int
}

func NewPos() *Pos {
	return &Pos{-1, -1}
}

// define Player object prototype
type Player struct {
	init *Pos
	curr *Pos
}

// initialize grid with -1
type Game struct {
	player_count int
	my_id        int
	my_direction string
	grid         [][]int
	player       []*Player
}

func NewGame() *Game {
	return &Game{
		player_count: -1,
		my_id:        -1,
		my_direction: "LEFT",
		grid:         new_matrix(30, 20, -1),
		player:       []*Player{},
	}
}

func simulate(sim_grid [][]int, scenario []string, xArr [][]int, yArr [][]int, turn int, totalScoreArr []int) []int {

	me := tron.my_id
	//my_player := tron.player[me]
	lostArr := []int{600, 600, 600, 600}
	lostArr[me] = 0
	if turn == 600 {
		return lostArr
	}

	// I move first
	xArr_next := make([][]int, tron.player_count)
	yArr_next := make([][]int, tron.player_count)
	scoreArr := []int{0, 0, 0, 0}
	for k := 0; k <= 3; k++ {
		player_id := (me + k) % 4
		if player_id >= tron.player_count {
			continue
		}
		score := 0
		x := xArr[player_id]
		y := yArr[player_id]
		len1 := len(x)
		x_next := make([]int, tron.player_count)
		y_next := make([]int, tron.player_count)
		for i := 0; i < len1; i++ {
			if x[i] >= 0 && x[i] <= 29 && y[i] >= 0 && y[i] <= 19 {
				// if we have scenarios, move along the scenario
				if player_id == me && scenario != nil {
					moveTo := move(x[i], y[i], scenario[0])
					//printErr('Current scenario: ' + scenario[0] + ", x[i]=" + x[i] + " y[i]=" + y[i] + ", moveTo=" + moveTo);
					moveToX := moveTo.X
					moveToY := moveTo.Y
					if moveToX >= 0 && moveToX <= 29 && moveToY >= 0 && moveToY <= 19 && sim_grid[moveToX][moveToY] < 0 {
						//if (turn===0)
						//    printErr("moveToX="+moveToX+",moveToY="+moveToY+",sim_grid="+sim_grid[moveToX][moveToY]);
						sim_grid[moveToX][moveToY] = me
						score += 1
						x_next = append(x_next, moveToX)
						y_next = append(y_next, moveToY)
					} else if turn == 0 {
						return lostArr
					}
					// otherwise move in all directions
				} else {
					if x[i] > 0 && sim_grid[x[i]-1][y[i]] < 0 {
						sim_grid[x[i]-1][y[i]] = player_id
						score += 1
						x_next = append(x_next, x[i]-1)
						y_next = append(y_next, y[i])
					}
					if y[i] > 0 && sim_grid[x[i]][y[i]-1] < 0 {
						sim_grid[x[i]][y[i]-1] = player_id
						score += 1
						x_next = append(x_next, x[i])
						y_next = append(y_next, y[i]-1)
					}
					if x[i] < 29 && sim_grid[x[i]+1][y[i]] < 0 {
						sim_grid[x[i]+1][y[i]] = player_id
						score += 1
						x_next = append(x_next, x[i]+1)
						y_next = append(y_next, y[i])
					}
					if y[i] < 19 && sim_grid[x[i]][y[i]+1] < 0 {
						sim_grid[x[i]][y[i]+1] = player_id
						score += 1
						x_next = append(x_next, x[i])
						y_next = append(y_next, y[i]+1)
					}
				} // if scenario != null
			} // if  good point
		} // for next length
		if score > 0 {
			xArr_next[player_id] = x_next
			yArr_next[player_id] = y_next
			totalScoreArr[player_id] += score
		} else {
			if RESUME_FLOODFILL_WHEN_PLAYER_DISAPPEARS {
				// the next array will stay as is until next round, unless the player disappears
				xArr_next[player_id] = xArr[player_id]
				yArr_next[player_id] = yArr[player_id]
			} else {
				xArr_next[player_id] = []int{}
				yArr_next[player_id] = []int{}
			}
			// if(player_id==3)
			//     printErr("SimTurn:" + turn + " Player " + player_id + " score=" + totalScoreArr[player_id]);
			if MAKE_LOSING_PLAYER_DISAPPEAR_FROM_SIM_GRID && (turn == 0 || totalScoreArr[player_id] > 0 && turn >= totalScoreArr[player_id]) {
				// if(player_id==1)
				//      printErr("SimTurn:" + turn + " Player " + player_id + " disappears.");
				matrix_overwrite(30, 20, sim_grid, player_id, -1)
				totalScoreArr[player_id] = 0
				xArr_next[player_id] = []int{}
				yArr_next[player_id] = []int{}
			}
		}
		scoreArr[player_id] = score
	} // for all players

	if scoreArr[0] == 0 && scoreArr[1] == 0 && scoreArr[2] == 0 && scoreArr[3] == 0 {
		return scoreArr
	}
	if len(scenario) > 1 {
		scenario = scenario[1:]
	} else {
		scenario = nil
	}
	next_turn := simulate(sim_grid, scenario, xArr_next, yArr_next, turn+1, totalScoreArr)
	return []int{scoreArr[0] + next_turn[0], scoreArr[1] + next_turn[1], scoreArr[2] + next_turn[2], scoreArr[3] + next_turn[3]}
}

func move(x int, y int, direction string) *Pos {
	switch direction {
	case "LEFT":
		return &Pos{x - 1, y}
	case "RIGHT":
		return &Pos{x + 1, y}
	case "UP":
		return &Pos{x, y - 1}
	case "DOWN":
		return &Pos{x, y + 1}
	}
	return nil
}

func score_eval_sumop_hug(score []int, neighborhood_penalty int) int {
	my_score := score[tron.my_id]
	sumop_score := 0
	if SCORE_EVAL_MAXOP {
		maxop_score := 0
		for j := 0; j <= 3; j++ {
			if j != tron.my_id && score[j] > maxop_score {
				maxop_score = score[j]
			}
		}
		sumop_score = maxop_score
	} else {
		for j := 0; j <= 3; j++ {
			if j != tron.my_id {
				sumop_score += score[j]
			}
		}
	}
	return my_score - sumop_score/SCORE_EVAL_SUMOP_DIVISOR - neighborhood_penalty/SCORE_EVAL_NEIGHBOR_DIVISOR
}

func check_neighbors(grid [][]int, moveTo *Pos, active_players int) int {
	penalty := 0
	x := moveTo.X
	y := moveTo.Y
	me := tron.my_id
	free_neighbors := 0
	if x >= 0 && x <= 29 && y >= 0 && y <= 19 {
		//penalize using up cells with many free neighbours
		if x > 0 && grid[x-1][y] < 0 {
			free_neighbors += 1
		}
		if y > 0 && grid[x][y-1] < 0 {
			free_neighbors += 1
		}
		if x < 29 && grid[x+1][y] < 0 {
			free_neighbors += 1
		}
		if y < 19 && grid[x][y+1] < 0 {
			free_neighbors += 1
		}
		penalty = free_neighbors

		if PENALIZE_OP && active_players > 2 && x > 0 && x < 29 && y > 0 && y > 29 {
			l := grid[x-1][y]
			lu := grid[x-1][y-1]
			u := grid[x][y-1]
			ru := grid[x+1][y-1]
			r := grid[x+1][y]
			rd := grid[x+1][y+1]
			d := grid[x][y+1]
			ld := grid[x-1][y+1]
			if l >= 0 && l != me || lu >= 0 && lu != me ||
				u >= 0 && u != me || ru >= 0 && ru != me ||
				r >= 0 && r != me || rd >= 0 && rd != me ||
				d >= 0 && d != me || ld >= 0 && ld != me {
				penalty += PENALIZE_OP_POINTS * SCORE_EVAL_NEIGHBOR_DIVISOR
			}
		}

		//penalize going to the edge cells too early
		if PENALIZE_EDGE && !PENALIZE_EDGE_WHEN_3_OR_4 || active_players > 2 && (x == 0 || x == 29 || y == 0 || y == 19) {
			penalty += PENALIZE_EDGE_POINTS * SCORE_EVAL_NEIGHBOR_DIVISOR
		}
	}
	return penalty
}

func generate_scenarios(depth int) [][]string {
	scenarios := [][]string{
		[]string{"LEFT"},
		[]string{"UP"},
		[]string{"RIGHT"},
		[]string{"DOWN"},
	}
	for i := 0; i < depth; i++ {
		scenarios = explode_scenarios(scenarios)
	}
	return scenarios
}

func explode_scenarios(scenarios [][]string) [][]string {
	exploded_scenarios := [][]string{}
	for i := 0; i < len(scenarios); i++ {
		for j := 0; j < 4; j++ {
			scenario := scenarios[i]
			new_scenario := append(scenario, directions[j])
			if !has_cycle(new_scenario) {
				exploded_scenarios = append(exploded_scenarios, new_scenario)
			}
		}
	}
	return exploded_scenarios
}

func has_cycle(scenario []string) bool {
	visited := make(map[int]bool)
	current := 0
	//printErr("checking cycle for scenario: " + scenario);
	for i := 0; i < len(scenario); i++ {
		direction := scenario[i]
		switch direction {
		case "LEFT":
			current -= 1
		case "RIGHT":
			current += 1
		case "UP":
			current -= 100
		case "DOWN":
			current += 100
		default:
		}
		//printErr(current);
		if _, ok := visited[current]; ok {
			return true
		}
		visited[current] = true
	}
	return false
}

// tests
// test_scenario=["LEFT","LEFT","LEFT","DOWN","RIGHT","RIGHT"];
// printErr(test_scenario+" has_cycle=" + has_cycle(test_scenario));
// test_scenario=["UP","LEFT","RIGHT","LEFT"];
// printErr(test_scenario+" has_cycle=" + has_cycle(test_scenario));

var (
	tron       = NewGame()
	directions = []string{"LEFT", "UP", "RIGHT", "DOWN"}
)

// create players
func main() {
	tron.player = make([]*Player, 4)
	for i := 0; i <= 3; i++ {
		tron.player[i] = &Player{}
		tron.player[i].init = NewPos()
		tron.player[i].curr = NewPos()
	}
	xArr := make([][]int, 4)
	yArr := make([][]int, 4)

	// game loop
	//totalTime := 0
	round := -1
	sim_grid := new_matrix(30, 20, -1)
	scenarios := generate_scenarios(SCENARIO_DEPTH)
	//total_score_arr := []int{0, 0, 0, 0}
	fmt.Fprintf(os.Stderr, "scenarios length=%d [%s],[%s]...[%s]", len(scenarios), scenarios[0], scenarios[1], scenarios[len(scenarios)-1])

	for {
		round += 1
		startTime := time.Now()
		fmt.Scan(&tron.player_count, &tron.my_id)
		active_players := 0
		for i := 0; i < tron.player_count; i++ {
			var X0, Y0, X1, Y1 int
			fmt.Scan(&X0, &Y0, &X1, &Y1)

			tron.player[i].init.X = X0 // starting X coordinate of lightcycle (or -1)
			tron.player[i].init.Y = Y0 // starting Y coordinate of lightcycle (or -1)
			if tron.player[i].init.X != -1 {
				active_players += 1
				tron.grid[tron.player[i].init.X][tron.player[i].init.Y] = i
				tron.player[i].curr.X = X1 // current X coordinate of lightcycle (can be the same as X0 if you play before this player)
				tron.player[i].curr.Y = Y1 // current Y coordinate of lightcycle (can be the same as Y0 if you play before this player)
				//printErr("Player " + i + "x=" + tron.player[i].curr.X + " y=" + tron.player[i].curr.Y);
				tron.grid[tron.player[i].curr.X][tron.player[i].curr.Y] = i
				xArr[i] = []int{tron.player[i].curr.X}
				yArr[i] = []int{tron.player[i].curr.Y}

			} else {
				matrix_overwrite(30, 20, tron.grid, i, -1)
				xArr[i] = []int{}
				yArr[i] = []int{}
			}
		}

		me := tron.player[tron.my_id]
		fmt.Fprintf(os.Stderr, "Me(%d):%d %d, active:%d", tron.my_id, me.curr.X, me.curr.Y, active_players)

		highest_score_eval := -9999
		var best_scenario []string
		for i := 0; i < len(scenarios); i++ {
			direction := scenarios[i][0]
			moveTo := move(me.curr.X, me.curr.Y, direction)
			matrix_copy_to(30, 20, sim_grid, tron.grid)

			total_score_arr := []int{0, 0, 0, 0}
			scenario_score := simulate(sim_grid, scenarios[i], xArr, yArr, 0, total_score_arr)
			neighborhood_penalty := check_neighbors(tron.grid, moveTo, active_players)
			score_eval := score_eval_sumop_hug(scenario_score, neighborhood_penalty)

			if len(DEBUG_SCENARIO) > 0 {
				ds_len := len(DEBUG_SCENARIO)
				ds_print := true
				for ds_i := 0; ds_i < ds_len; ds_i++ {
					if scenarios[i][ds_i] != DEBUG_SCENARIO[ds_i] {
						ds_print = false
						break
					}
				}
				if ds_print {
					fmt.Fprintf(os.Stderr, "Scenario: %s %d (%d) eval: %d", scenarios[i], scenario_score, neighborhood_penalty, score_eval)
				}
			}
			if score_eval > highest_score_eval {
				fmt.Fprintf(os.Stderr, "New highest score: %s %d (%d) eval: %d", scenarios[i], scenario_score, neighborhood_penalty, score_eval)
				highest_score_eval = score_eval
				tron.my_direction = direction
				best_scenario = scenarios[i]
			}
		}

		fmt.Fprintf(os.Stderr, "Highest eval %s: %d", best_scenario, highest_score_eval)
		elapsed := time.Since(startTime)
		fmt.Fprintf(os.Stderr, "Elapsed: %v", elapsed)
		//	    if round!=0{
		//	        totalElapsed+= elapsed
		//	        //printErr('Avg Time: ' + Math.round(totalTime/round) + ' ms');
		//	        fmt.Fprintf("Total elapsed: %v",totalElapsed)
		//	    }
		fmt.Println(tron.my_direction) // A single line with UP, DOWN, LEFT or RIGHT
	}
}

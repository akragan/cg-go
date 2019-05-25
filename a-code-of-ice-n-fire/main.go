package main

import "fmt"
import "sort"
import "os"

//import "bufio"
//import "strings"

const (
	GridDim = 12

	IdMe   = 0
	IdOp   = 1
	IdVoid = -1

	CmdWait       = 0
	CmdMove       = 1
	CmdTrain      = 2
	CmdBuildMine  = 3
	CmdBuildTower = 4

	TypeHq    = 0
	TypeMine  = 1
	TypeTower = 2

	CostTrain1 = 10
	CostTrain2 = 20
	CostTrain3 = 30

	CostKeep1 = 1
	CostKeep2 = 4
	CostKeep3 = 20

	CostMine1 = 20
	CostTower = 15

	CellVoid    = '#'
	CellNeutral = '.'
	CellMine    = '$'

	RowNeutral = "............"

	CellMeA  = 'O' // my active cell
	CellMeNA = 'o' // my inactive cell
	CellMeH  = 'H' // my HQ
	CellMeM  = 'M' // my active Mine
	CellMeNM = 'N' // my iNactive Mine
	CellMeT  = 'T' // my active Tower
	CellMeNT = 'F' // my iNactive Tower
	CellMeP  = 'P' // my Tower-protected cell

	CellOpA  = 'X' // op active cell
	CellOpNA = 'x' // op inactive cell
	CellOpH  = 'h' // op HQ
	CellOpM  = 'm' // op active mine
	CellOpNM = 'n' // op Not active Mine
	CellOpT  = 't' // op active Tower
	CellOpNT = 'f' // op iNactive Tower
	CellOpP  = 'p' // op Tower-protected cell

	CellMeU  = 'U' // my unit level 1
	CellMeU2 = 'K' // my unit level 2
	CellMeU3 = 'G' // my unit level 3

	CellOpU  = 'u' // op unit level 1
	CellOpU2 = 'k' // op unit level 2
	CellOpU3 = 'g' // op unit level 3

	DirLeft  = 0
	DirUp    = 1
	DirRight = 2
	DirDown  = 3

	Min1 = 3
	Min2 = 2

	InfDist = 100
)

var (
	g = &Game{}

	DirDRUL = []int{DirDown, DirRight, DirUp, DirLeft}
	DirLURD = []int{DirLeft, DirUp, DirRight, DirDown}
)

type PositionQueue []*Position

func (s PositionQueue) Put(v *Position) PositionQueue {
	return append(s, v)
}

func (s PositionQueue) TakeFirst() (PositionQueue, *Position) {
	return s[1:], s[0]
}

func (s PositionQueue) IsEmpty() bool {
	return len(s) == 0
}

func distance(x1 int, y1 int, x2 int, y2 int) int {
	dist := 0
	if x1 > x2 {
		dist += x1 - x2
	} else {
		dist += x2 - x1
	}
	if y1 > y2 {
		dist += y1 - y2
	} else {
		dist += y2 - y1
	}
	return dist
}

type HasPosition interface {
	Pos() *Position
}

type Position struct {
	X    int
	Y    int
	Dist int
}

func (this *Position) Pos() *Position {
	return this
}

func (this *Position) toInt() int {
	return this.X*100 + this.Y
}

func (this *Position) fromInt(i int) {
	this.X = i / 100
	this.Y = i % 100
}

func (this *Position) setDistance(other *Position) int {
	this.Dist = distance(this.X, this.Y, other.X, other.Y)
	return this.Dist
}

// unsafe
func (this *Position) getCell(grid [][]rune) rune {
	return grid[this.Y][this.X]
}

func (this *Position) setCell(grid [][]rune, cell rune) {
	grid[this.Y][this.X] = cell
}

func (this *Position) getIntCell(intGrid [][]int) int {
	return intGrid[this.Y][this.X]
}

func (this *Position) setIntCell(intGrid [][]int, cell int) {
	intGrid[this.Y][this.X] = cell
}

func (this *Position) set(x int, y int) *Position {
	this.X = x
	this.Y = y
	return this
}

func (this *Position) neighbour(direction int) *Position {
	var n *Position
	switch direction {
	case DirLeft:
		if this.X > 0 {
			n = &Position{X: this.X - 1, Y: this.Y}
		}
	case DirRight:
		if this.X < GridDim-1 {
			n = &Position{X: this.X + 1, Y: this.Y}
		}
	case DirUp:
		if this.Y > 0 {
			n = &Position{X: this.X, Y: this.Y - 1}
		}
	case DirDown:
		if this.Y < GridDim-1 {
			n = &Position{X: this.X, Y: this.Y + 1}
		}
	}
	return n
}

type Player struct {
	Id     int
	Gold   int
	Income int

	NbUnits  int
	NbUnits1 int
	NbUnits2 int
	NbUnits3 int

	NbMines   int
	NbTowers  int
	MineSpots []*Position

	ActiveArea int
	Upkeep     int

	MinUnitDistGoal int
	MinDistGoal     *Position
}

func (this *Player) addUnit(u *Unit) {
	this.NbUnits++
	switch u.Level {
	case 1:
		this.NbUnits1++
		this.Upkeep += CostKeep1
	case 2:
		this.NbUnits2++
		this.Upkeep += CostKeep2
	case 3:
		this.NbUnits3++
		this.Upkeep += CostKeep3
	}
}

func (this *Player) income() int {
	return this.ActiveArea + 4*this.NbMines - this.Upkeep
}

func (this *Player) mineCost() int {
	return CostMine1 + this.NbMines*4
}

type Building struct {
	Type  int
	Owner int
	X     int
	Y     int
}

func (this *Building) Pos() *Position {
	return &Position{X: this.X, Y: this.Y}
}

type Unit struct {
	Id    int
	X     int
	Y     int
	Owner int
	Level int
}

func (this *Unit) Pos() *Position {
	return &Position{X: this.X, Y: this.Y}
}

type Command struct {
	Type int
	Info int
	X    int
	Y    int
}

func (this *Command) Pos() *Position {
	return &Position{X: this.X, Y: this.Y}
}

type Game struct {
	NbMines     int
	Mines       []*Position
	MineGrid    [][]rune
	HqMe        *Position
	HqOp        *Position
	InitNeutral int
	DistGrid    [][]int
}

func initGame() {
	fmt.Scan(&g.NbMines)
	g.InitNeutral = 0
	g.Mines = make([]*Position, g.NbMines)
	g.MineGrid = make([][]rune, GridDim)
	g.DistGrid = make([][]int, GridDim)
	for i := 0; i < GridDim; i++ {
		g.MineGrid[i] = []rune(RowNeutral)
		g.DistGrid[i] = []int{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}
	}
	for i := 0; i < g.NbMines; i++ {
		mine := &Position{}
		fmt.Scan(&mine.X, &mine.Y)
		g.Mines[i] = mine
		mine.setCell(g.MineGrid, CellMine)
	}
}

func initDistGrid(grid [][]rune) {
	pos := &Position{X: g.HqOp.X, Y: g.HqOp.Y, Dist: 0}
	todo := PositionQueue{pos}
	for !todo.IsEmpty() {
		todo, pos = todo.TakeFirst()
		if pos.getIntCell(g.DistGrid) != -1 {
			continue
		}
		//fmt.Fprintf(os.Stderr, "init DistGrid: (%d,%d):%d, queue size=%d\n", pos.X, pos.Y, pos.Dist, len(todo))
		if pos.getCell(grid) == CellVoid {
			pos.setIntCell(g.DistGrid, InfDist)
		} else {
			pos.setIntCell(g.DistGrid, pos.Dist)
			for _, dir := range DirDRUL {
				nbrPos := pos.neighbour(dir)
				if nbrPos != nil && nbrPos.getIntCell(g.DistGrid) == -1 {
					nbrPos.Dist = pos.Dist + 1
					todo = todo.Put(nbrPos)
					//fmt.Fprintf(os.Stderr, "\tdir=%v add (%d,%d):%d, queue size=%d\n", dir, nbrPos.X, nbrPos.Y, nbrPos.Dist, len(todo))
				} // if -1 (Dist not set)
			} // for all dirs
		} // if/else cell void
		//printDistGrid()
	} // for queue non-empty
	printDistGrid()
}

func printDistGrid() {
	for i := 0; i < GridDim; i++ {
		line := ""
		for j := 0; j < GridDim; j++ {
			line += fmt.Sprintf("%d ", g.DistGrid[i][j])
		}
		fmt.Fprintf(os.Stderr, "%v\n", line)
	}
}

type State struct {
	Me          *Player
	Op          *Player
	Grid        [][]rune
	Neutral     int
	NeutralPct  float32
	NbBuildings int
	Buildings   []*Building
	NbUnits     int
	Units       []*Unit
	UnitGrid    [][]rune

	Commands []*Command
}

func (s *State) init(turn int) {
	pos := &Position{}

	s.Me = &Player{}
	s.Me.MinUnitDistGoal = InfDist
	s.Me.MinDistGoal = &Position{X: -1, Y: -1, Dist: InfDist}
	fmt.Scan(&s.Me.Gold)
	fmt.Scan(&s.Me.Income)

	s.Op = &Player{}
	s.Op.MinUnitDistGoal = InfDist
	fmt.Scan(&s.Op.Gold)
	fmt.Scan(&s.Op.Income)

	s.Grid = make([][]rune, GridDim)
	s.UnitGrid = make([][]rune, GridDim)
	s.Neutral = 0
	for i := 0; i < GridDim; i++ {
		var line string
		fmt.Scan(&line)
		//fmt.Fprintf(os.Stderr, "%v\n", line)
		s.Grid[i] = []rune(line)
		for j := 0; j < GridDim; j++ {
			pos.set(j, i)
			if line[j] == CellMeA {
				s.Me.ActiveArea++
				if turn > 0 {
					dist := pos.getIntCell(g.DistGrid)
					if dist < s.Me.MinDistGoal.Dist {
						s.Me.MinDistGoal.set(j, i).Dist = dist
					}
				}
			} else if line[j] == CellOpA {
				s.Op.ActiveArea++
			} else if line[j] == CellNeutral {
				if turn == 0 {
					g.InitNeutral += 1
				}
				s.Neutral += 1
			}
		}
		s.UnitGrid[i] = []rune(RowNeutral)
	}
	s.NeutralPct = float32(s.Neutral) / float32(g.InitNeutral)
	fmt.Fprintf(os.Stderr, "%d: NeutralPct=%v\n", turn, s.NeutralPct)
	fmt.Fprintf(os.Stderr, "%d: Me.MinDistGoal=(%d,%d):%d\n", turn, s.Me.MinDistGoal.X, s.Me.MinDistGoal.Y, s.Me.MinDistGoal.Dist)
	fmt.Fprintf(os.Stderr, "%d: TrainChainWin:%v Gold:%d TrainChainCost=%d\n", turn, s.Me.Gold >= s.Me.MinDistGoal.Dist*CostTrain1, s.Me.Gold, s.Me.MinDistGoal.Dist*CostTrain1)

	fmt.Scan(&s.NbBuildings)
	s.Buildings = make([]*Building, s.NbBuildings)
	for i := 0; i < s.NbBuildings; i++ {
		b := Building{}
		fmt.Scan(&b.Owner, &b.Type, &b.X, &b.Y)
		s.Buildings[i] = &b
		bPos := b.Pos()
		switch b.Type {
		case TypeHq:
			if b.Owner == IdMe {
				g.HqMe = bPos
				bPos.setCell(s.Grid, CellMeH)
			} else {
				g.HqOp = bPos
				bPos.setCell(s.Grid, CellOpH)
			}
		case TypeMine:
			if b.Owner == IdMe {
				if bPos.getCell(s.Grid) == CellMeA {
					bPos.setCell(s.Grid, CellMeM)
				} else {
					bPos.setCell(s.Grid, CellMeNM)
				}
				// TODO find out if inactive mines count towards building cost
				s.Me.NbMines++
			} else {
				if bPos.getCell(s.Grid) == CellOpA {
					bPos.setCell(s.Grid, CellOpM)
				} else {
					bPos.setCell(s.Grid, CellOpNM)
				}
				s.Op.NbMines++
			}
		case TypeTower:
			if b.Owner == IdMe {
				if bPos.getCell(s.Grid) == CellMeA {
					bPos.setCell(s.Grid, CellMeT)
				} else {
					bPos.setCell(s.Grid, CellMeNT)
				}
				s.Me.NbTowers++
			} else {
				if bPos.getCell(s.Grid) == CellOpA {
					bPos.setCell(s.Grid, CellOpT)
					// set Op tower-protected cells
					for _, dir := range DirDRUL {
						nbrPos := bPos.neighbour(dir)
						if nbrPos != nil && nbrPos.getCell(s.Grid) == CellOpA {
							nbrPos.setCell(s.Grid, CellOpP)
						}
					}
				} else {
					bPos.setCell(s.Grid, CellOpNT)
				}
				s.Op.NbTowers++
			}
		}
	}

	if turn == 0 {
		initDistGrid(s.Grid)
	}

	fmt.Scan(&s.NbUnits)
	s.Units = make([]*Unit, s.NbUnits)
	for i := 0; i < s.NbUnits; i++ {
		u := &Unit{}
		fmt.Scan(&u.Owner, &u.Id, &u.Level, &u.X, &u.Y)
		s.Units[i] = u
		pos.set(u.X, u.Y)

		if u.Owner == IdMe {
			pos.setDistance(g.HqOp)
			if s.Me.MinUnitDistGoal > pos.Dist {
				s.Me.MinUnitDistGoal = pos.Dist
			}
			switch u.Level {
			case 1:
				pos.setCell(s.UnitGrid, CellMeU)
			case 2:
				pos.setCell(s.UnitGrid, CellMeU2)
			case 3:
				pos.setCell(s.UnitGrid, CellMeU3)
			}
			s.Me.addUnit(u)
		} else {
			pos.setDistance(g.HqMe)
			if s.Op.MinUnitDistGoal > pos.Dist {
				s.Op.MinUnitDistGoal = pos.Dist
			}
			switch u.Level {
			case 1:
				pos.setCell(s.UnitGrid, CellOpU)
			case 2:
				pos.setCell(s.UnitGrid, CellOpU2)
			case 3:
				pos.setCell(s.UnitGrid, CellOpU3)
			}
			s.Op.addUnit(u)
		}
	}

	s.Commands = []*Command{&Command{Type: CmdWait}}
}

func (s *State) addBuildMine(at *Position) {
	s.Commands = append(s.Commands, &Command{Type: CmdBuildMine, X: at.X, Y: at.Y})
	at.setCell(s.Grid, CellMeM)
	s.Me.Gold -= s.Me.mineCost()
	s.Me.NbMines += 1
}

func (s *State) addBuildTower(at *Position) {
	s.Commands = append(s.Commands, &Command{Type: CmdBuildTower, X: at.X, Y: at.Y})
	at.setCell(s.Grid, CellMeT)
	s.Me.Gold -= CostTower
	s.Me.NbTowers += 1
}

func (s *State) addMove(u *Unit, from *Position, to *Position) {
	s.Commands = append(s.Commands, &Command{Type: CmdMove, Info: u.Id, X: to.X, Y: to.Y})
	to.setCell(s.Grid, CellMeA)
	to.setCell(s.UnitGrid, CellMeU)
	from.setCell(s.UnitGrid, CellNeutral)
}

func (s *State) addTrain(at *Position, level int) {
	s.Commands = append(s.Commands, &Command{Type: CmdTrain, Info: level, X: at.X, Y: at.Y})
	at.setCell(s.UnitGrid, CellMeU)
	switch level {
	case 1:
		s.Me.Gold -= CostTrain1
	case 2:
		s.Me.Gold -= CostTrain2
	case 3:
		s.Me.Gold -= CostTrain3
	}
	s.Me.addUnit(&Unit{IdMe, -1, level, at.X, at.Y})
}

func (s *State) action() string {
	cmdsStr := ""
	for i := 0; i < len(s.Commands); i++ {
		if i > 0 {
			cmdsStr += ";"
		}
		cmd := s.Commands[i]
		switch cmd.Type {
		case CmdWait:
			cmdsStr += "WAIT"
		case CmdTrain:
			cmdsStr += fmt.Sprintf("TRAIN %d %d %d", cmd.Info, cmd.X, cmd.Y)
		case CmdMove:
			cmdsStr += fmt.Sprintf("MOVE %d %d %d", cmd.Info, cmd.X, cmd.Y)
		case CmdBuildMine:
			cmdsStr += fmt.Sprintf("BUILD MINE %d %d", cmd.X, cmd.Y)
		case CmdBuildTower:
			cmdsStr += fmt.Sprintf("BUILD TOWER %d %d", cmd.X, cmd.Y)
		}
	}
	cmdsStr += fmt.Sprintf(";MSG A:%d U:%d I:%d", s.Me.ActiveArea, s.Me.NbUnits, s.Me.income())
	return cmdsStr
}

type CommandSelector struct {
	Candidates []*CandidateCommand
}

type CandidateCommand struct {
	Unit  *Unit     // Move
	From  *Position // Move
	To    *Position // Move, Train
	Level int       // Train
	Value int       // Move, Train
}

func (this *CommandSelector) appendMove(u *Unit, from *Position, to *Position, value int) {
	this.Candidates = append(this.Candidates, &CandidateCommand{
		Unit:  u,
		From:  from,
		To:    to,
		Value: value,
	})
}

func (this *CommandSelector) appendTrain(level int, at *Position, value int) {
	this.Candidates = append(this.Candidates, &CandidateCommand{
		To:    at,
		Level: level,
		Value: value,
	})
}

func (this *CommandSelector) sort() {
	if len(this.Candidates) < 2 {
		return
	}
	sort.Slice(this.Candidates, func(i, j int) bool { return this.Candidates[i].Value > this.Candidates[j].Value })
}

func (this *CommandSelector) best() *CandidateCommand {
	if len(this.Candidates) == 0 {
		return nil
	}
	this.sort()
	return this.Candidates[0]
}

func myUnitCell(cell rune) bool {
	return cell == CellMeU || cell == CellMeU2 || cell == CellMeU3
}

func opUnitCell(cell rune) bool {
	return cell == CellOpU || cell == CellOpU2 || cell == CellOpU3
}

func anyUnitCell(cell rune) bool {
	return myUnitCell(cell) || opUnitCell(cell)
}

func myActiveCell(cell rune) bool {
	return cell == CellMeA || cell == CellMeH || cell == CellMeM || cell == CellMeT || cell == CellMeP
}

func opActiveCell(cell rune) bool {
	return cell == CellOpA || cell == CellOpH || cell == CellOpM || cell == CellOpT || cell == CellOpP
}

func compactFactor(pos *Position, grid [][]rune) int {
	count := 0
	for _, dir := range DirDRUL {
		nbrPos := pos.neighbour(dir)
		if nbrPos != nil {
			nbrCell := nbrPos.getCell(grid)
			if myActiveCell(nbrCell) {
				count += 1
			}
		}
	}
	return count
}

func isWedge(pos *Position, grid [][]rune) bool {
	lPos := pos.neighbour(DirLeft)
	lOpA := lPos != nil && opActiveCell(lPos.getCell(grid))

	rPos := pos.neighbour(DirRight)
	rOpA := rPos != nil && opActiveCell(rPos.getCell(grid))

	uPos := pos.neighbour(DirUp)
	uOpA := uPos != nil && opActiveCell(uPos.getCell(grid))

	dPos := pos.neighbour(DirDown)
	dOpA := dPos != nil && opActiveCell(dPos.getCell(grid))

	return lOpA && rOpA && !uOpA && !dOpA || !lOpA && !rOpA && uOpA && dOpA
}

func moveUnits(s *State) {
	pos := &Position{}
	dirs := DirDRUL
	if g.HqMe.X != 0 {
		dirs = DirLURD
	}
	for i := 0; i < s.NbUnits; i++ {
		u := s.Units[i]
		if u.Owner != IdMe || u.Id == -1 { // -1 for newly trained units that cannot move
			continue
		}
		pos.set(u.X, u.Y)
		//fmt.Fprintf(os.Stderr, "Unit: %d Pos: %d %d HQ: %d %d \n", u.Id, pos.X, pos.Y, g.HqMe.X, g.HqMe.Y)
		candidateCmds := &CommandSelector{}
		for _, dir := range dirs {

			nbrPos := pos.neighbour(dir)

			if nbrPos == nil {
				continue
			}

			nbrCell := nbrPos.getCell(s.Grid)
			unitCell := nbrPos.getCell(s.UnitGrid)

			if nbrCell == CellVoid {
				continue
			}
			// Op HQ capturing moves (by any unit)
			if nbrCell == CellOpH {
				candidateCmds.appendMove(u, pos, nbrPos, 20)
				continue
			}
			// Op active TOWER capturing moves (only by l3 unit)
			if u.Level == 3 && unitCell == CellOpP && !myUnitCell(unitCell) {
				candidateCmds.appendMove(u, pos, nbrPos, 19)
				continue
			}
			// Op TOWER-protected land capturing moves (only by l3 unit)
			if u.Level == 3 && unitCell == CellOpT && !myUnitCell(unitCell) {
				candidateCmds.appendMove(u, pos, nbrPos, 18)
				continue
			}
			// Op inactive TOWER capturing moves (only by l3 unit)
			if u.Level == 3 && unitCell == CellOpNT && !myUnitCell(unitCell) {
				candidateCmds.appendMove(u, pos, nbrPos, 17)
				continue
			}
			// Op unit l3 capturing moves (only by l3 unit)
			if u.Level == 3 && unitCell == CellOpU3 && !myUnitCell(unitCell) {
				candidateCmds.appendMove(u, pos, nbrPos, 16)
				continue
			}
			// Op unit l2 capturing moves (only by l3 unit)
			if u.Level == 3 && unitCell == CellOpU2 && !myUnitCell(unitCell) {
				candidateCmds.appendMove(u, pos, nbrPos, 15)
				continue
			}
			// Op active MINE capturing moves (by any unit)
			if nbrCell == CellOpM && !myUnitCell(unitCell) {
				candidateCmds.appendMove(u, pos, nbrPos, 14)
				continue
			}
			// Op unit l1 capturing moves (only by any l2 or l3 unit)
			if (u.Level == 3 || u.Level == 2) && unitCell == CellOpU && !myUnitCell(unitCell) {
				candidateCmds.appendMove(u, pos, nbrPos, 13)
				//s.addMove(u, pos, nbrPos)
				continue
			}
			// Op INactive MINE capturing moves (by any unit)
			if nbrCell == CellOpNM && !myUnitCell(unitCell) {
				candidateCmds.appendMove(u, pos, nbrPos, 12)
				continue
			}
			// Op active land capturing moves (by any unit)
			// ++ priority for cells splitting Op territory
			// + priority for cells keeping my territory compact
			if nbrCell == CellOpA && !anyUnitCell(unitCell) {
				if isWedge(nbrPos, s.Grid) {
					candidateCmds.appendMove(u, pos, nbrPos, 11)
				} else if compactFactor(nbrPos, s.Grid) > 1 {
					candidateCmds.appendMove(u, pos, nbrPos, 10)
				} else {
					candidateCmds.appendMove(u, pos, nbrPos, 9)
				}
				continue
			}
			// Op INactive land capturing moves (by any unit)
			// + more priority for cells keeping my territory compact
			if nbrCell == CellOpNA && !myUnitCell(unitCell) {
				if compactFactor(nbrPos, s.Grid) > 1 {
					candidateCmds.appendMove(u, pos, nbrPos, 8)
				} else {
					candidateCmds.appendMove(u, pos, nbrPos, 7)
				}
				continue
			}
			// new land capturing moves (by any unit)
			// + more priority for cells keeping my territory compact
			if nbrCell == CellNeutral && !myUnitCell(unitCell) {
				if compactFactor(nbrPos, s.Grid) > 1 {
					candidateCmds.appendMove(u, pos, nbrPos, 5)
				} else {
					candidateCmds.appendMove(u, pos, nbrPos, 4)
				}
				continue
			}
			// standing my ground if faced with uncapturable enemy (lvl2 right now)
			// i.e. issuing invalid move command on purpose
			if u.Level == 2 && unitCell == CellOpU2 {
				candidateCmds.appendMove(u, pos, pos, 1)
				continue
			}

			// moving to another free cell (by any unit)
			// value depends on whether we're getting closer or further from Op Hq
			// 1 if closer, 0 if same, -1 if further
			if nbrCell == CellMeA && !myUnitCell(unitCell) {
				currDist := pos.getIntCell(g.DistGrid)
				nbrDist := nbrPos.getIntCell(g.DistGrid)
				candidateCmds.appendMove(u, pos, nbrPos, currDist-nbrDist)
				continue
			}
		} //for dir
		// pick the best move for unit
		if bestCmd := candidateCmds.best(); bestCmd != nil {
			fmt.Fprintf(os.Stderr, "Unit:%d, Candidates:%d, Best:%d X:%d Y:%d\n", bestCmd.Unit.Id, len(candidateCmds.Candidates), bestCmd.Value, bestCmd.To.X, bestCmd.To.Y)
			s.addMove(bestCmd.Unit, bestCmd.From, bestCmd.To)
		}
	}
}

func trainUnitInNeighbourhood(cmds *CommandSelector, s *State, pos *Position, dirs []int, cellType rune, highVal int) {
	for _, dir := range dirs {
		nbrPos := pos.neighbour(dir)
		if nbrPos != nil {
			nbrCell := nbrPos.getCell(s.Grid)
			unitCell := nbrPos.getCell(s.UnitGrid)
			if nbrCell == cellType && unitCell == CellNeutral {
				if (s.Me.NbUnits < Min1 || s.NeutralPct > 0.2) &&
					s.Me.Gold > CostTrain1 && s.Me.Gold < 2*CostTrain2 {
					cmds.appendTrain(1, nbrPos, highVal)
				} else if s.Me.NbUnits >= s.Op.NbUnits &&
					s.Me.income() > 2*CostKeep3 &&
					s.Me.Gold > CostTrain3 {
					cmds.appendTrain(3, nbrPos, highVal)
				} else if s.Me.income() > 2*CostKeep2 &&
					s.Me.Gold > CostTrain2 {
					cmds.appendTrain(2, nbrPos, highVal)
				} else {
					// found a cell but couldn't train
					return
				}
			}
		}
	} //for dir
}

func trainUnits(s *State) {
	pos := &Position{}
	candidateCmds := &CommandSelector{}

	dirs := DirLURD
	if g.HqMe.X != 0 {
		dirs = DirDRUL
	}

	// train in new areas
	for j := 0; j < GridDim; j++ {
		for i := 0; i < GridDim; i++ {
			if s.Me.Gold < CostTrain1 {
				// no gold to train any units
				return
			}
			pos.set(i, j)
			cell := pos.getCell(s.Grid)
			if cell != CellMeA && cell != CellMeH && cell != CellMeM {
				continue
			}
			trainUnitInNeighbourhood(candidateCmds, s, pos, dirs, CellNeutral, 12)
		} // for i
	} // for j

	if len(candidateCmds.Candidates) == 0 {
		// train in areas neighbouring inactive Op cells
		for j := 0; j < GridDim; j++ {
			for i := 0; i < GridDim; i++ {
				if s.Me.Gold < CostTrain1 {
					// no gold to train any units
					return
				}
				pos.set(i, j)
				cell := pos.getCell(s.Grid)
				if cell != CellOpNA && cell != CellOpNM && cell != CellOpNT {
					continue
				}
				trainUnitInNeighbourhood(candidateCmds, s, pos, dirs, CellMeA, 9)
			} // for i
		} // for j
	}

	if len(candidateCmds.Candidates) == 0 {
		// train in areas neighbouring active Op cells
		for j := 0; j < GridDim; j++ {
			for i := 0; i < GridDim; i++ {
				if s.Me.Gold < CostTrain1 {
					// no gold to train any units
					return
				}
				pos.set(i, j)
				cell := pos.getCell(s.Grid)
				if cell != CellOpA && cell != CellOpM && cell != CellOpT && cell != CellOpP {
					continue
				}
				trainUnitInNeighbourhood(candidateCmds, s, pos, dirs, CellMeA, 6)
			} // for i
		} // for j
	}

	if len(candidateCmds.Candidates) == 0 {
		// worst case train at headquarters
		pos.set(g.HqMe.X, g.HqMe.Y)
		trainUnitInNeighbourhood(candidateCmds, s, pos, dirs, CellMeA, 3)
	}
	// sort and execute
	candidateCmds.sort()
	for _, cmd := range candidateCmds.Candidates {
		cost := costTrain(cmd.Level)
		if cost < s.Me.Gold {
			s.addTrain(cmd.To, cmd.Level)
		}
	}
}

func costTrain(level int) int {
	switch level {
	case 2:
		return CostTrain2
	case 3:
		return CostTrain3
	}
	return CostTrain1
}

func buildMinesAndTowers(s *State) {
	if s.Me.NbUnits >= Min1 {
		// try building 1 mine
		if s.Op.NbMines > 0 && s.Me.NbMines == 0 && s.Me.Gold > s.Me.mineCost() {
			if g.HqMe.X == 0 {
				// build mine at (1,0)
				pos := &Position{X: 1, Y: 0}
				if pos.getCell(s.Grid) == CellMeA && pos.getCell(s.UnitGrid) == CellNeutral {
					s.addBuildMine(pos)
				}
			} else {
				// build mine at (10,11)
				pos := &Position{X: 10, Y: 11}
				if pos.getCell(s.Grid) == CellMeA && pos.getCell(s.UnitGrid) == CellNeutral {
					s.addBuildMine(pos)
				}
			}
		}
		// try building 1 tower
		if s.Op.MinUnitDistGoal <= 5 && s.Me.NbTowers == 0 && s.Me.Gold > CostTower {
			if g.HqMe.X == 0 {
				// build tower at (1,1)
				pos := &Position{X: 1, Y: 1}
				if pos.getCell(s.Grid) == CellMeA && pos.getCell(s.UnitGrid) == CellNeutral {
					s.addBuildTower(pos)
				}
			} else {
				// build tower at (10,10)
				pos := &Position{X: 10, Y: 10}
				if pos.getCell(s.Grid) == CellMeA && pos.getCell(s.UnitGrid) == CellNeutral {
					s.addBuildTower(pos)
				}
			}
		}
	}
}

func main() {
	initGame()
	for i := 0; ; i++ {
		s := &State{}
		s.init(i)

		// generate candidate commands (start with WAIT that never hurts)

		// 0. look for BUILD MINE and/or TOWER commands
		buildMinesAndTowers(s)

		// 1. look at MOVE commands
		moveUnits(s)

		// 2. look at TRAIN commands
		trainUnits(s)

		// fmt.Fprintln(os.Stderr, "Debug messages...")
		fmt.Println(s.action()) // Write action to stdout
	} // for
}

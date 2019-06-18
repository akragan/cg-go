package main

import "fmt"
import "sort"
import "os"
import "time"
import "math/rand"
import "math"

//import "bufio"
//import "strings"

const (
	// debug
	DebugActiveArea = false
	DebugTrain      = false

	//options

	StandGroundL1 = true
	StandGroundL2 = true

	SortUnitsAsc  = true
	SortUnitsDesc = false // used only if SortUnitsAsc==false

	MaxTowersInTouch = 2
	MaxTowers        = 6
	Min1             = 3
	//Min2      = 2

	MoveBackwards                   = false
	RandomDirsAtInitDistGrid        = false
	AbortTrainCmdsOnNegativeEvalChg = true
	TrainNegativeEvalPainTolerance  = -25.0

	EvalDiscountRate                    = 5.0
	EvalHqCaptureFactor                 = 100.0
	EvalExpectedIncomeFactorFromNeutral = 0.25

	//constants
	GridDim = 12

	IdMe   = 0
	IdOp   = 1
	IdVoid = -1

	CmdAbort      = -1 // to remove Train commands
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

	CostMine  = 15
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

	InfDist = 100
)

var (
	g = &Game{}

	DirDRUL = []int{DirDown, DirRight, DirUp, DirLeft}
	DirRDLU = []int{DirRight, DirDown, DirLeft, DirUp}

	DirLURD = []int{DirLeft, DirUp, DirRight, DirDown}
	DirULDR = []int{DirUp, DirLeft, DirDown, DirRight}
)

//---------------------------------------------------------------------------------------
//---------------------------------------------------------------------------------------

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

func (this *Position) set(x int, y int) *Position {
	this.X = x
	this.Y = y
	return this
}

func (this *Position) sameAs(other *Position) bool {
	return this.X == other.X && this.Y == other.Y
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

func (this *Position) onEdge(grid [][]rune) bool {
	for _, dir := range DirDRUL {
		nbrPos := this.neighbour(dir)
		if nbrPos == nil || nbrPos.getCell(grid) == CellVoid {
			return true
		}
	}
	return false
}

func (this *Position) freedom(grid [][]rune) int {
	f := 4
	for _, dir := range DirDRUL {
		nbrPos := this.neighbour(dir)
		if nbrPos == nil || nbrPos.getCell(grid) == CellVoid {
			f -= 1
		}
	}
	return f
}

func (this *Position) findNeighbour(grid [][]rune, cell rune) int {
	for _, dir := range DirDRUL {
		nbrPos := this.neighbour(dir)
		if nbrPos != nil && nbrPos.getCell(grid) == cell {
			return dir
		}
	}
	return -1
}

func (this *Position) isOrHasNeighbour(grid [][]rune, cell rune) bool {
	if this.getCell(grid) == cell {
		return true
	}
	for _, dir := range DirDRUL {
		nbrPos := this.neighbour(dir)
		if nbrPos != nil && nbrPos.getCell(grid) == cell {
			return true
		}
	}
	return false
}

func (this *Position) isOrHasNeighbourAtDist2(grid [][]rune, cell rune) bool {
	if this.getCell(grid) == cell {
		return true
	}
	for _, dir1 := range DirDRUL {
		nbrPos := this.neighbour(dir1)
		if nbrPos == nil {
			continue
		}
		if nbrPos.getCell(grid) == cell {
			return true
		}
		// neighbour dist 2
		for _, dir2 := range DirDRUL {
			nbrPos2 := nbrPos.neighbour(dir2)
			if nbrPos2 != nil && nbrPos2.getCell(grid) == cell {
				return true
			}
		}
	}
	return false
}

func (this *Position) findNeighbourDir(distGrid [][]int, dist int) int {
	for _, dir := range DirDRUL {
		nbrPos := this.neighbour(dir)
		if nbrPos != nil && nbrPos.getIntCell(distGrid) == dist {
			return dir
		}
	}
	return -1
}

func (this *Position) neighbour(direction int) *Position {
	switch direction {
	case DirLeft:
		if this.X > 0 {
			return &Position{X: this.X - 1, Y: this.Y}
		}
	case DirRight:
		if this.X < GridDim-1 {
			return &Position{X: this.X + 1, Y: this.Y}
		}
	case DirUp:
		if this.Y > 0 {
			return &Position{X: this.X, Y: this.Y - 1}
		}
	case DirDown:
		if this.Y < GridDim-1 {
			return &Position{X: this.X, Y: this.Y + 1}
		}
	}
	return nil
}

func (this *Position) diagonalNeighbour(dir1 int, dir2 int) *Position {
	switch {
	case dir1 == DirLeft && dir2 == DirUp || dir2 == DirLeft && dir1 == DirUp:
		if this.X > 0 && this.Y > 0 {
			return &Position{X: this.X - 1, Y: this.Y - 1}
		}
	case dir1 == DirRight && dir2 == DirUp || dir2 == DirRight && dir1 == DirUp:
		if this.X < GridDim-1 && this.Y > 0 {
			return &Position{X: this.X + 1, Y: this.Y - 1}
		}
	case dir1 == DirLeft && dir2 == DirDown || dir2 == DirLeft && dir2 == DirDown:
		if this.X > 0 && this.Y < GridDim-1 {
			return &Position{X: this.X - 1, Y: this.Y + 1}
		}
	case dir1 == DirRight && dir2 == DirDown || dir2 == DirRight && dir1 == DirDown:
		if this.X < GridDim-1 && this.Y < GridDim-1 {
			return &Position{X: this.X + 1, Y: this.Y + 1}
		}
	}
	return nil
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
	Id      int
	X       int
	Y       int
	Owner   int
	Level   int
	Freedom int
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

//---------------------------------------------------------------------------------------
//---------------------------------------------------------------------------------------

type Player struct {
	Id     int
	Game   *GamePlayer
	State  *State
	Other  *Player
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

	MinUnitDistGoal   int
	MinDistGoal       *Position
	MinDistGoalUnit   *Unit
	ChainTrainWin     bool
	ChainTrainWinNext bool

	MinChainTrainWinCost    int
	ActualChainTrainWinCost int
	RoundsToHqCapture       float64
	MilitaryPower           int
	ExpectedMilitaryPower   int
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

func (this *Player) addActiveArea(pos *Position) {
	this.ActiveArea++
	var dist int
	if this.Game.Initialized {
		dist = pos.getIntCell(this.Game.DistGrid)
	} else {
		dist = 22
	}
	if dist < this.MinDistGoal.Dist {
		this.MinDistGoal.set(pos.X, pos.Y).Dist = dist
	}

}

func (this *Player) income() int {
	return this.ActiveArea + 4*this.NbMines - this.Upkeep
}

func (this *Player) expectedIncome() int {
	pctUnits := float64(this.NbUnits+1) / float64(this.State.NbUnits+2)
	return this.ActiveArea +
		int(EvalExpectedIncomeFactorFromNeutral*pctUnits*float64(this.State.Neutral)) +
		4*this.NbMines - this.Upkeep
}

func (p *Player) isMyUnit(unitCell rune) bool {
	if p.Id == IdMe {
		return myUnitCell(unitCell)
	}
	return opUnitCell(unitCell)
}

func (p *Player) isEnemyUnitLevel1(unitCell rune) bool {
	if p.Id == IdMe {
		return unitCell == CellOpU
	}
	return unitCell == CellMeU
}

func (p *Player) isEnemyUnitLevel2or3(unitCell rune) bool {
	if p.Id == IdMe {
		return unitCell == CellOpU2 || unitCell == CellOpU3
	}
	return unitCell == CellMeU2 || unitCell == CellMeU3
}

func (p *Player) isMyActiveCell(cell rune) bool {
	if p.Id == IdMe {
		return myActiveCell(cell)
	}
	return opActiveCell(cell)
}

func (p *Player) isEnemyTowerOrProtected(cell rune) bool {
	if p.Id == IdMe {
		return cell == CellOpT || cell == CellOpP
	}
	return cell == CellMeT || cell == CellMeP
}

func (p *Player) calculateChainTrainWin(moveFirst bool, execute bool) {
	fmt.Fprintf(os.Stderr, "%d: [%s] Calculating ChainTrainWin: Gold=%d MinTrainChainCost=%d\n", g.Turn, p.Game.Name, p.Gold, p.MinDistGoal.Dist*CostTrain1)
	pos := p.MinDistGoal
	unitCell := pos.getCell(p.State.UnitGrid)
	isMyUnit := p.isMyUnit(unitCell)
	posDist := pos.getIntCell(p.Game.DistGrid)
	actualCost := 0
	cmds := &CommandSelector{}
	//fmt.Fprintf(os.Stderr, "start loop\n")
	for posDist != 0 {
		dir := pos.getIntCell(p.Game.DirGrid)
		//fmt.Fprintf(os.Stderr, "(%d,%d): posDist=%d dir=%d\n", pos.X, pos.Y, posDist, dir)
		fromPos := pos
		pos = pos.neighbour(dir)
		posDist = pos.getIntCell(p.Game.DistGrid)
		cell := pos.getCell(p.State.Grid)
		unitCell := pos.getCell(p.State.UnitGrid)
		level := 1
		if p.isEnemyUnitLevel1(unitCell) {
			level = 2
		}
		if p.isEnemyTowerOrProtected(cell) || p.isEnemyUnitLevel2or3(unitCell) {
			level = 3
		}
		if moveFirst && isMyUnit && level == 1 { // fix to account for more free first moves of level 2 and 3
			// first move for free
			fmt.Fprintf(os.Stderr, "\t[%s] using free move first to move to (%d,%d) level=%d\n", p.Game.Name, pos.X, pos.Y, level)
			// add move command
			if execute {
				cmds.appendMove(p.MinDistGoalUnit, fromPos, pos, posDist)
			}
		} else {
			actualCost += costTrain(level)
			if execute {
				cmds.appendTrain(level, pos, posDist)
			}
		}
		moveFirst = false
	}
	p.ActualChainTrainWinCost = actualCost
	//fmt.Fprintf(os.Stderr, "end loop\n")
	if p.Gold < actualCost {
		fmt.Fprintf(os.Stderr, "\t[%s] Abort: Gold=%d ActualCost=%d\n", p.Game.Name, p.Gold, actualCost)
		return
	}
	fmt.Fprintf(os.Stderr, "\t[%s] Proceed: Gold=%d ActualCost=%d\n", p.Game.Name, p.Gold, actualCost)
	if !execute {
		return
	}
	for i, cmd := range cmds.Candidates {
		if cmd.Level == 0 { //move command
			p.State.addMove(cmd.Unit, cmd.From, cmd.To)
			fmt.Fprintf(os.Stderr, "\t%d: value %d, move %d to (%d,%d)\n", i, cmd.Value, cmd.Unit.Id, cmd.To.X, cmd.To.Y)
		} else {
			p.State.addTrain(cmd.To, cmd.Level)
			fmt.Fprintf(os.Stderr, "\t%d: value %d, level %d at (%d,%d)\n", i, cmd.Value, cmd.Level, cmd.To.X, cmd.To.Y)
		}
	}
}

func (p *Player) evaluate() {
	p.RoundsToHqCapture = 100.0
	if p.ActualChainTrainWinCost < p.Gold {
		p.RoundsToHqCapture = 0.0
	} else if p.income() > 0 {
		p.RoundsToHqCapture = float64(p.ActualChainTrainWinCost-p.Gold) / float64(p.expectedIncome())
	}

	p.MilitaryPower = p.NbUnits3*CostTrain3 + p.NbUnits2*CostTrain2 + p.NbUnits1*CostTrain1 + p.Gold
	p.ExpectedMilitaryPower = p.MilitaryPower + (100-g.Turn)*p.income()
	if p.ExpectedMilitaryPower < 0 {
		p.ExpectedMilitaryPower = 0
	}

}

//---------------------------------------------------------------------------------------
//---------------------------------------------------------------------------------------

type GamePlayer struct {
	Name        string
	Hq          *Position
	Other       *GamePlayer
	DistGrid    [][]int
	DirGrid     [][]int
	Initialized bool
}

func (this *Player) recalculateActiveArea() {
	activeCells := make([][]rune, GridDim)
	for i := 0; i < GridDim; i++ {
		activeCells[i] = []rune(RowNeutral)
	}
	activeArea := 0
	pos := &Position{X: this.Game.Other.Hq.X, Y: this.Game.Other.Hq.Y, Dist: 0}
	todo := PositionQueue{pos}
	for !todo.IsEmpty() {
		todo, pos = todo.TakeFirst()
		activeCell := pos.getCell(activeCells)
		if activeCell != CellNeutral {
			continue
		}
		cell := pos.getCell(this.State.Grid)
		if this.isMyActiveCell(cell) {
			activeArea += 1
			pos.setCell(activeCells, CellMine)
		} else {
			pos.setCell(activeCells, CellVoid)
		}
		dirs := DirDRUL
		for _, dir := range dirs {
			nbrPos := pos.neighbour(dir)
			if nbrPos != nil && nbrPos.getCell(activeCells) == CellNeutral {
				todo = todo.Put(nbrPos)
			} // if not visited
		} // for all dirs
	}
	activeAreaChg := activeArea - this.ActiveArea
	if activeAreaChg != 0 {
		fmt.Fprintf(os.Stderr, "%d active area changed by %d (from %d to %d)\n", this.Id, activeAreaChg, this.ActiveArea, activeArea)
		//this.ActiveArea = activeArea
		//TODO update active area
		//this.updateActive(activeCells)
	} else {
		if DebugActiveArea {
			fmt.Fprintf(os.Stderr, "%d active area unchanged (%d)\n", this.Id, this.ActiveArea)
		}
	}
}

func (this *GamePlayer) initDistGrid(grid [][]rune) {
	pos := &Position{X: this.Other.Hq.X, Y: this.Other.Hq.Y, Dist: 0}
	todo := PositionQueue{pos}
	for !todo.IsEmpty() {
		todo, pos = todo.TakeFirst()
		if pos.getIntCell(this.DistGrid) != -1 {
			continue
		}
		//fmt.Fprintf(os.Stderr, "init DistGrid: (%d,%d):%d, queue size=%d\n", pos.X, pos.Y, pos.Dist, len(todo))
		if pos.getCell(grid) == CellVoid {
			pos.setIntCell(this.DistGrid, InfDist)
		} else {
			pos.setIntCell(this.DistGrid, pos.Dist)
			if pos.Dist != 0 {
				pos.setIntCell(this.DirGrid, pos.findNeighbourDir(this.DistGrid, pos.Dist-1))
			}
			dirs := DirDRUL
			if RandomDirsAtInitDistGrid {
				dirs = randDirs()
			}
			for _, dir := range dirs {
				nbrPos := pos.neighbour(dir)
				if nbrPos != nil && nbrPos.getIntCell(this.DistGrid) == -1 {
					nbrPos.Dist = pos.Dist + 1
					todo = todo.Put(nbrPos)
					//fmt.Fprintf(os.Stderr, "\tdir=%v add (%d,%d):%d, queue size=%d\n", dir, nbrPos.X, nbrPos.Y, nbrPos.Dist, len(todo))
				} // if -1 (Dist not set)
			} // for all dirs
		} // if/else cell void
		//printDistGrid()
	} // for queue non-empty
	this.Initialized = true
	printIntGrid(this.Name+" DistGrid", this.DistGrid)
	printIntGrid(this.Name+" DirGrid", this.DirGrid)
}

func printIntGrid(label string, grid [][]int) {
	fmt.Fprintf(os.Stderr, "%s:\n", label)
	for i := 0; i < GridDim; i++ {
		line := ""
		for j := 0; j < GridDim; j++ {
			line += fmt.Sprintf("%d ", grid[i][j])
		}
		fmt.Fprintf(os.Stderr, "%v\n", line)
	}
}

//---------------------------------------------------------------------------------------
//---------------------------------------------------------------------------------------

type Game struct {
	Turn           int
	DiscountFactor float64

	TurnTime time.Time
	RespTime time.Time

	Me *GamePlayer
	Op *GamePlayer

	NbMines     int
	Mines       []*Position
	MineGrid    [][]rune
	InitNeutral int
	InTouch     bool
}

func (g *Game) nextTurn() {
	g.Turn += 1
	g.DiscountFactor = math.Exp((float64(g.Turn)/100.0 - 1.0) * EvalDiscountRate)
}

func initGame() {
	g.Turn = 0
	g.DiscountFactor = math.Exp(-1.0 * EvalDiscountRate)

	fmt.Scan(&g.NbMines)
	g.InitNeutral = 0
	g.InTouch = false
	g.Mines = make([]*Position, g.NbMines)
	g.MineGrid = make([][]rune, GridDim)

	g.Me = &GamePlayer{Name: "Me", Initialized: false}
	g.Me.DistGrid = make([][]int, GridDim)
	g.Me.DirGrid = make([][]int, GridDim)

	g.Op = &GamePlayer{Name: "Op", Initialized: false}
	g.Op.DistGrid = make([][]int, GridDim)
	g.Op.DirGrid = make([][]int, GridDim)

	g.Me.Other = g.Op
	g.Op.Other = g.Me

	for i := 0; i < GridDim; i++ {
		g.MineGrid[i] = []rune(RowNeutral)
		g.Me.DistGrid[i] = []int{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}
		g.Me.DirGrid[i] = []int{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}
		g.Op.DistGrid[i] = []int{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}
		g.Op.DirGrid[i] = []int{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}
	}
	for i := 0; i < g.NbMines; i++ {
		mine := &Position{}
		fmt.Scan(&mine.X, &mine.Y)
		g.Mines[i] = mine
		mine.setCell(g.MineGrid, CellMine)
	}
}

//---------------------------------------------------------------------------------------
//---------------------------------------------------------------------------------------

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
	UnitById    map[int]*Unit
	UnitGrid    [][]rune
	// my commands to action
	Commands []*Command
	// state eval
	MilitaryPowerEval float64
	HqCaptureEval     float64
	Eval              float64
}

func (s *State) evaluate(label string) {
	fmt.Fprintf(os.Stderr, "%d: evaluating state %s\n", g.Turn, label)

	s.Me.evaluate()
	s.Op.evaluate()

	s.HqCaptureEval = EvalHqCaptureFactor * (s.Op.RoundsToHqCapture - s.Me.RoundsToHqCapture) * (1 - g.DiscountFactor)
	s.MilitaryPowerEval = float64(s.Me.ExpectedMilitaryPower-s.Op.ExpectedMilitaryPower) * g.DiscountFactor
	s.Eval = s.HqCaptureEval + s.MilitaryPowerEval

	fmt.Fprintf(os.Stderr, "\tHQ capture eval=%.1f\tMeTurnsToHQ=%.1f OpTurnsToHQ=%.1f MeIncome=%d->%d OpIncome=%d->%d\n",
		s.HqCaptureEval, s.Me.RoundsToHqCapture, s.Op.RoundsToHqCapture, s.Me.income(), s.Me.expectedIncome(), s.Op.income(), s.Op.expectedIncome())
	fmt.Fprintf(os.Stderr, "\tmilitary eval=%.1f\tMeExMP=%v OpExMP=%v\n", s.MilitaryPowerEval, s.Me.ExpectedMilitaryPower, s.Op.ExpectedMilitaryPower)
	fmt.Fprintf(os.Stderr, "%d: eval=%.1f\t(df=%.2f)\n", g.Turn, s.Eval, g.DiscountFactor)
}

func (s *State) init() {
	pos := &Position{}

	s.Me = &Player{Id: IdMe, Game: g.Me, State: s}
	s.Me.MinUnitDistGoal = InfDist
	s.Me.MinDistGoal = &Position{X: -1, Y: -1, Dist: InfDist}
	fmt.Scan(&s.Me.Gold)
	fmt.Scan(&s.Me.Income)

	s.Op = &Player{Id: IdOp, Game: g.Op, State: s}
	s.Op.MinUnitDistGoal = InfDist
	s.Op.MinDistGoal = &Position{X: -1, Y: -1, Dist: InfDist}
	fmt.Scan(&s.Op.Gold)
	fmt.Scan(&s.Op.Income)

	s.Me.Other = s.Op
	s.Op.Other = s.Me

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
				s.Me.addActiveArea(pos)
			} else if line[j] == CellOpA {
				s.Op.addActiveArea(pos)
			} else if line[j] == CellNeutral {
				if g.Turn == 0 {
					g.InitNeutral += 1
				}
				s.Neutral += 1
			}
		}
		s.UnitGrid[i] = []rune(RowNeutral)
	}
	s.NeutralPct = float32(s.Neutral) / float32(g.InitNeutral)
	s.Me.MinChainTrainWinCost = s.Me.MinDistGoal.Dist * CostTrain1
	s.Me.ActualChainTrainWinCost = s.Me.MinChainTrainWinCost
	s.Me.ChainTrainWin = s.Me.Gold >= s.Me.MinDistGoal.Dist*CostTrain1
	s.Me.ChainTrainWinNext = s.Me.Gold+s.Me.income() >= (s.Me.MinDistGoal.Dist-1)*CostTrain1
	s.Op.MinChainTrainWinCost = s.Op.MinDistGoal.Dist * CostTrain1
	s.Op.ActualChainTrainWinCost = s.Op.MinChainTrainWinCost
	s.Op.ChainTrainWin = s.Op.Gold >= s.Op.MinDistGoal.Dist*CostTrain1
	s.Op.ChainTrainWinNext = s.Op.Gold+s.Op.income() >= (s.Op.MinDistGoal.Dist-1)*CostTrain1

	fmt.Fprintf(os.Stderr, "%d: NeutralPct=%v\n", g.Turn, s.NeutralPct)
	fmt.Fprintf(os.Stderr, "%d: Me.MinDistGoal=(%d,%d):%d\n", g.Turn, s.Me.MinDistGoal.X, s.Me.MinDistGoal.Y, s.Me.MinDistGoal.Dist)
	fmt.Fprintf(os.Stderr, "%d: Me.ChainTrainWin:%v Next:%v Gold:%d TrainChainCost=%d\n", g.Turn, s.Me.ChainTrainWin, s.Me.ChainTrainWinNext, s.Me.Gold, s.Me.MinDistGoal.Dist*CostTrain1)
	fmt.Fprintf(os.Stderr, "%d: Op.MinDistGoal=(%d,%d):%d\n", g.Turn, s.Op.MinDistGoal.X, s.Op.MinDistGoal.Y, s.Op.MinDistGoal.Dist)
	fmt.Fprintf(os.Stderr, "%d: Op.ChainTrainWin:%v Next:%v Gold:%d TrainChainCost=%d\n", g.Turn, s.Op.ChainTrainWin, s.Op.ChainTrainWinNext, s.Op.Gold, s.Op.MinDistGoal.Dist*CostTrain1)

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
				g.Me.Hq = bPos
				bPos.setCell(s.Grid, CellMeH)
			} else {
				g.Op.Hq = bPos
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
			cell := bPos.getCell(s.Grid)
			if b.Owner == IdMe {
				// tower cell active or protected by another tower
				if cell == CellMeA || cell == CellMeP {
					bPos.setCell(s.Grid, CellMeT)
				} else {
					bPos.setCell(s.Grid, CellMeNT)
				}
				s.Me.NbTowers++
			} else {
				// tower cell active or protected by another tower
				if cell == CellOpA || cell == CellOpP {
					bPos.setCell(s.Grid, CellOpT)
					// set Op tower-protected cells
					for _, dir := range DirDRUL {
						nbrPos := bPos.neighbour(dir)
						if nbrPos != nil {
							nbrCell := nbrPos.getCell(s.Grid)
							if nbrCell == CellOpA || nbrCell == CellOpM || nbrCell == CellOpH {
								nbrPos.setCell(s.Grid, CellOpP)
							}
						}
					}
				} else {
					bPos.setCell(s.Grid, CellOpNT)
				}
				s.Op.NbTowers++
			}
		}
	}

	if g.Turn == 0 {
		g.Me.initDistGrid(s.Grid)
		g.Op.initDistGrid(s.Grid)
	}

	fmt.Scan(&s.NbUnits)
	s.Units = make([]*Unit, s.NbUnits)
	s.UnitById = make(map[int]*Unit)
	for i := 0; i < s.NbUnits; i++ {
		u := &Unit{}
		fmt.Scan(&u.Owner, &u.Id, &u.Level, &u.X, &u.Y)
		s.Units[i] = u
		s.UnitById[u.Id] = u
		pos.set(u.X, u.Y)
		if u.Owner == IdMe {
			if !g.InTouch {
				if pos.findNeighbour(s.Grid, CellOpA) != -1 {
					g.InTouch = true
				}
			}
			u.Freedom = pos.freedom(s.Grid)
			pos.setDistance(g.Op.Hq)
			if s.Me.MinUnitDistGoal > pos.Dist {
				s.Me.MinUnitDistGoal = pos.Dist
			}
			if s.Me.MinDistGoal.sameAs(pos) {
				s.Me.MinDistGoalUnit = u
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
			pos.setDistance(g.Me.Hq)
			if s.Op.MinUnitDistGoal > pos.Dist {
				s.Op.MinUnitDistGoal = pos.Dist
			}
			if s.Op.MinDistGoal.sameAs(pos) {
				s.Op.MinDistGoalUnit = u
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
	// sort units from l1 to l3 (l1 will move first)
	// the idea being for them to be moving into enemy's camp first)
	// then by freedom - units less free should move first
	if SortUnitsAsc {
		sort.Slice(s.Units, func(i, j int) bool {
			if s.Units[i].Level == s.Units[j].Level {
				return s.Units[i].Freedom < s.Units[j].Freedom
			}
			return s.Units[i].Level < s.Units[j].Level
		})
	} else if SortUnitsDesc {
		sort.Slice(s.Units, func(i, j int) bool {
			if s.Units[i].Level == s.Units[j].Level {
				return s.Units[i].Freedom < s.Units[j].Freedom
			}
			return s.Units[i].Level > s.Units[j].Level
		})
	}
	g.RespTime = time.Now()
	s.Commands = []*Command{&Command{Type: CmdWait}}
}

func (s *State) addBuildMine(at *Position) {
	s.Commands = append(s.Commands, &Command{Type: CmdBuildMine, X: at.X, Y: at.Y})
	at.setCell(s.Grid, CellMeM)
	s.Me.Gold -= CostMine
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
	if !from.sameAs(to) {
		to.setCell(s.UnitGrid, CellMeU)
		from.setCell(s.UnitGrid, CellNeutral)

		cell := to.getCell(s.Grid)
		if !myActiveCell(cell) {
			to.setCell(s.Grid, CellMeA)
			s.Me.addActiveArea(to)
		}
	}
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
	s.Me.addUnit(&Unit{Owner: IdMe, Id: -1, Level: level, X: at.X, Y: at.Y})
	cell := at.getCell(s.Grid)
	if !myActiveCell(cell) {
		at.setCell(s.Grid, CellMeA)
		s.Me.addActiveArea(at)
	}
}

func (s *State) action() string {
	cmdsStr := ""
	for i := 0; i < len(s.Commands); i++ {
		cmd := s.Commands[i]
		switch cmd.Type {
		case CmdWait:
			cmdsStr += "WAIT"
		case CmdTrain:
			cmdsStr += fmt.Sprintf(";TRAIN %d %d %d", cmd.Info, cmd.X, cmd.Y)
		case CmdMove:
			cmdsStr += fmt.Sprintf(";MOVE %d %d %d", cmd.Info, cmd.X, cmd.Y)
		case CmdBuildMine:
			cmdsStr += fmt.Sprintf(";BUILD MINE %d %d", cmd.X, cmd.Y)
		case CmdBuildTower:
			cmdsStr += fmt.Sprintf(";BUILD TOWER %d %d", cmd.X, cmd.Y)
		}
	}
	cmdsStr += fmt.Sprintf(";MSG Eval:%.1f", s.Eval)
	return cmdsStr
}

//---------------------------------------------------------------------------------------
//---------------------------------------------------------------------------------------

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

func (this *CommandSelector) dedupe() {
	// dedupe by setting level to 0 to remove
	intSet := make(map[int]bool)
	for _, cmd := range this.Candidates {
		intPos := cmd.To.toInt()
		_, dupe := intSet[intPos]
		if dupe {
			cmd.Level = 0
		} else {
			intSet[intPos] = true
		}
	}
}

func (this *CommandSelector) best() *CandidateCommand {
	if len(this.Candidates) == 0 {
		return nil
	}
	this.sort()
	return this.Candidates[0]
}

//---------------------------------------------------------------------------------------
//---------------------------------------------------------------------------------------

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

func boolRand() bool {
	return rand.Intn(2) == 0
}

func randDirs() []int {
	r := boolRand()
	switch {
	case r && g.Me.Hq.X == 0:
		return DirDRUL
	case !r && g.Me.Hq.X == 0:
		return DirRDLU
	case r && g.Me.Hq.X != 0:
		return DirLURD
	default:
		return DirULDR
	}
}

//---------------------------------------------------------------------------------------
//---------------------------------------------------------------------------------------

func moveUnits(s *State) {
	pos := &Position{}
	for i := 0; i < s.NbUnits; i++ {
		u := s.Units[i]
		if u.Owner != IdMe || u.Id == -1 { // -1 for newly trained units that cannot move
			continue
		}
		pos.set(u.X, u.Y)
		//fmt.Fprintf(os.Stderr, "Unit: %d Pos: %d %d HQ: %d %d \n", u.Id, pos.X, pos.Y, g.HqMe.X, g.HqMe.Y)
		dirs := randDirs()
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
			if u.Level == 3 && nbrCell == CellOpT && !myUnitCell(unitCell) {
				candidateCmds.appendMove(u, pos, nbrPos, 19)
				continue
			}
			// Op TOWER-protected land capturing moves (only by l3 unit)
			if u.Level == 3 && nbrCell == CellOpP && !myUnitCell(unitCell) {
				candidateCmds.appendMove(u, pos, nbrPos, 18)
				continue
			}
			// Op inactive TOWER capturing moves (only by l3 unit)
			if u.Level == 3 && nbrCell == CellOpNT && !myUnitCell(unitCell) {
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
			if (u.Level == 3 || u.Level == 2) && unitCell == CellOpU && nbrCell != CellOpP && !myUnitCell(unitCell) {
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
			// standing my ground if faced with uncapturable enemy (lvl 1 and 2)
			// i.e. issuing invalid move command on purpose
			if StandGroundL2 && u.Level == 2 && unitCell == CellOpU2 ||
				StandGroundL1 && u.Level == 1 && unitCell == CellOpU {
				candidateCmds.appendMove(u, pos, pos, 0)
				continue
			}

			// just moving to another free cell (by any unit)
			// value depends on whether we're getting closer or further from Op Hq
			// 1 if closer, 0 if same, -1 if further
			if nbrCell == CellMeA && !myUnitCell(unitCell) {
				currDist := pos.getIntCell(g.Me.DistGrid)
				nbrDist := nbrPos.getIntCell(g.Me.DistGrid)
				if currDist-nbrDist >= 0 || MoveBackwards {
					candidateCmds.appendMove(u, pos, nbrPos, currDist-nbrDist)
				}
				continue
			}
		} //for dir
		// pick the best move for unit
		if bestCmd := candidateCmds.best(); bestCmd != nil {
			//fmt.Fprintf(os.Stderr, "Unit:%d, Candidates:%d, Best:%d X:%d Y:%d\n", bestCmd.Unit.Id, len(candidateCmds.Candidates), bestCmd.Value, bestCmd.To.X, bestCmd.To.Y)
			s.addMove(bestCmd.Unit, bestCmd.From, bestCmd.To)
		}
	}
}

//---------------------------------------------------------------------------------------
//---------------------------------------------------------------------------------------
// this produces dupe candidate train commands (in the same spots)
// as cells are neighbours of several other cells
// needs to be sorted and de-duped before execution
func candidateTrainCmdsInNeighbourhood(cmds *CommandSelector, s *State, pos *Position) {

	// 1. consider current cell (lowest value)
	cell := pos.getCell(s.Grid)
	unitCell := pos.getCell(s.UnitGrid)

	if cell == CellMeA && unitCell == CellNeutral {
		// copy pos
		pos := &Position{X: pos.X, Y: pos.Y}
		// consider level 1
		if (s.Me.NbUnits < Min1 || s.NeutralPct > 0.2) &&
			s.Me.Gold > CostTrain1 && s.Me.Gold < 2*CostTrain2 {
			cmds.appendTrain(1, pos, 3-pos.getIntCell(g.Me.DistGrid))
		}
		// consider level 2
		if s.Me.NbUnits < 5*s.Op.NbUnits &&
			s.Me.income() > 2*CostKeep2 &&
			s.Me.Gold > CostTrain2 {
			cmds.appendTrain(2, pos, 1-pos.getIntCell(g.Me.DistGrid))
		}
		// consider level 3
		if (s.Me.NbUnits >= s.Op.NbUnits || s.Me.NbUnits3 == 0 && s.Op.NbUnits3 > 0) &&
			s.Me.income() > 2*CostKeep3 &&
			s.Me.Gold > CostTrain3 {
			cmds.appendTrain(3, pos, 2-pos.getIntCell(g.Me.DistGrid))
		}
	}

	// 2. consider neighbourhood (greater value)
	dirs := randDirs()
	for _, dir := range dirs {
		nbrPos := pos.neighbour(dir)
		if nbrPos == nil {
			continue
		}
		nbrCell := nbrPos.getCell(s.Grid)
		if myActiveCell(nbrCell) {
			// will be considered in its own right
			continue
		}
		nbrUnitCell := nbrPos.getCell(s.UnitGrid)
		bonus := 0
		if isWedge(nbrPos, s.Grid) {
			bonus += 10
		}

		if (nbrCell == CellNeutral || nbrCell == CellOpNA || nbrCell == CellOpNM || nbrCell == CellOpNT) &&
			nbrUnitCell == CellNeutral {
			// consider level 1
			if (s.Me.NbUnits < Min1 || s.NeutralPct > 0.2) &&
				s.Me.Gold > CostTrain1 && s.Me.Gold < 3*CostTrain2 {
				cmds.appendTrain(1, nbrPos, 6+bonus)
			}
			// consider level 2
			if s.Me.NbUnits < 5*s.Op.NbUnits &&
				s.Me.income() > 2*CostKeep2 &&
				s.Me.Gold > CostTrain2 {
				cmds.appendTrain(2, nbrPos, 4+bonus)
			}
			// consider level 3
			if (s.Me.NbUnits >= s.Op.NbUnits || s.Me.NbUnits3 == 0 && s.Op.NbUnits3 > 0) &&
				s.Me.income() > 2*CostKeep3 &&
				s.Me.Gold > CostTrain3 {
				cmds.appendTrain(3, nbrPos, 5+bonus)
			}
		}

		if (nbrCell == CellOpA || nbrCell == CellOpM) && nbrUnitCell == CellNeutral {
			// consider level 1
			if (s.Me.NbUnits < Min1 || s.NeutralPct > 0.2) &&
				s.Me.Gold > CostTrain1 && s.Me.Gold < 2*CostTrain2 {
				cmds.appendTrain(1, nbrPos, 9+bonus)
			}
			// consider level 2
			if (s.Me.NbUnits < 5*s.Op.NbUnits) &&
				s.Me.income() > 2*CostKeep2 &&
				s.Me.Gold > CostTrain2 {
				cmds.appendTrain(2, nbrPos, 8+bonus)
			}
			// consider level 3
			if (s.Me.NbUnits >= s.Op.NbUnits || s.Me.NbUnits3 == 0 && s.Op.NbUnits3 > 0) &&
				s.Me.income() > 2*CostKeep3 &&
				s.Me.Gold > CostTrain3 {
				cmds.appendTrain(3, nbrPos, 7+bonus)
			}
		}

		if nbrUnitCell == CellOpU && nbrCell != CellOpP {
			// consider level 2 and 3
			if (s.Me.NbUnits < 5*s.Op.NbUnits) &&
				s.Me.income() > 2*CostKeep2 &&
				s.Me.Gold > CostTrain2 {
				cmds.appendTrain(2, nbrPos, 11+bonus)
			}
			// consider level 3
			if (s.Me.NbUnits >= s.Op.NbUnits || s.Me.NbUnits3 == 0 && s.Op.NbUnits3 > 0) &&
				s.Me.income() > 2*CostKeep3 &&
				s.Me.Gold > CostTrain3 {
				cmds.appendTrain(3, nbrPos, 10+bonus)
			}
		}

		if nbrUnitCell == CellOpU2 {
			// consider level 3
			if (s.Me.NbUnits >= s.Op.NbUnits || s.Me.NbUnits3 == 0 && s.Op.NbUnits3 > 0) &&
				s.Me.income() > 2*CostKeep3 &&
				s.Me.Gold > CostTrain3 {
				cmds.appendTrain(3, nbrPos, 12+bonus)
			}
		}

		if nbrCell == CellOpT || nbrCell == CellOpP {
			// consider level 3
			if (s.Me.NbUnits >= s.Op.NbUnits || s.Me.NbUnits3 == 0 && s.Op.NbUnits3 > 0) &&
				s.Me.income() > 2*CostKeep3 &&
				s.Me.Gold > CostTrain3 {
				cmds.appendTrain(3, nbrPos, 13+bonus)
			}
		}

		if nbrUnitCell == CellOpU3 {
			// consider level 3
			if (!nbrPos.isOrHasNeighbourAtDist2(s.UnitGrid, CellMeU3) || s.Me.NbUnits3 == 0 && s.Op.NbUnits3 > 0) &&
				s.Me.income() > 2*CostKeep3 &&
				s.Me.Gold > CostTrain3 {
				cmds.appendTrain(3, nbrPos, 15+bonus)
			}
		}

		if nbrCell == CellOpP && nbrPos.sameAs(g.Op.Hq) {
			// consider level 3
			if s.Me.Gold > CostTrain3 {
				cmds.appendTrain(3, nbrPos, 100)
			}
		}

		if nbrCell == CellOpH {
			// consider level 1
			if s.Me.Gold > CostTrain1 {
				cmds.appendTrain(1, nbrPos, 100)
			}
			// consider level 2
			if s.Me.Gold > CostTrain2 {
				cmds.appendTrain(2, nbrPos, 100)
			}
			// consider level 3
			if s.Me.Gold > CostTrain3 {
				cmds.appendTrain(3, nbrPos, 100)
			}
		}

	} //for dir
}

func trainUnits(s *State) *CommandSelector {
	if s.Me.Gold < CostTrain1 {
		// no gold to train any units
		return nil
	}
	pos := &Position{}
	candidateCmds := &CommandSelector{}
	for j := 0; j < GridDim; j++ {
		for i := 0; i < GridDim; i++ {
			pos.set(i, j)
			cell := pos.getCell(s.Grid)
			if !myActiveCell(cell) {
				// can only train on and next to active area
				continue
			}
			candidateTrainCmdsInNeighbourhood(candidateCmds, s, pos)
		} // for i
	} // for j

	// sort, dedupe and execute
	candidateCmds.sort()
	candidateCmds.dedupe()
	return candidateCmds
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

//---------------------------------------------------------------------------------------
//---------------------------------------------------------------------------------------

func getHqTowerPosition() *Position {
	if g.Me.Hq.X == 0 {
		// build tower at (1,1)
		pos := &Position{X: 1, Y: 1}
		if pos.getCell(g.MineGrid) == CellMine {
			// if (1,1) is a mine, try (0,1) instead
			pos = &Position{X: 0, Y: 1}
		}
		return pos
	}
	// else build tower at (10,10)
	pos := &Position{X: 10, Y: 10}
	if pos.getCell(g.MineGrid) == CellMine {
		// if (10,10) is a mine, try (11,10) instead
		pos = &Position{X: 11, Y: 10}
	}
	return pos
}

func getHqMinePosition() *Position {
	if g.Me.Hq.X == 0 {
		// build mine at (1,0)
		return &Position{X: 1, Y: 0}
	}
	// else build mine at (10,11)
	return &Position{X: 10, Y: 11}
}

func findTowerSpotBeyondDist2(s *State, pos *Position) *Position {
	for (pos.getCell(s.Grid) != CellMeA ||
		pos.getCell(s.UnitGrid) != CellNeutral ||
		pos.getCell(g.MineGrid) == CellMine ||
		pos.isOrHasNeighbourAtDist2(s.Grid, CellMeT) ||
		pos.isOrHasNeighbourAtDist2(s.Grid, CellMeNT)) &&
		!pos.sameAs(g.Me.Hq) {
		//fmt.Fprintf(os.Stderr, "\t traversing (%d,%d)\n", pos.X, pos.Y)
		pos = pos.neighbour(pos.getIntCell(g.Op.DirGrid))
	}
	fmt.Fprintf(os.Stderr, "%d: Tower candidate at (%d,%d)\n", g.Turn, pos.X, pos.Y)
	if !pos.sameAs(g.Me.Hq) {
		fmt.Fprintf(os.Stderr, "\t accepted (%d,%d)\n", pos.X, pos.Y)
		return pos
	}
	fmt.Fprintf(os.Stderr, "\t rejected (%d,%d)\n", pos.X, pos.Y)
	return nil
}

func findTowerSpotBeyondDist1(s *State, pos *Position) *Position {
	for (pos.getCell(s.Grid) != CellMeA ||
		pos.getCell(s.UnitGrid) != CellNeutral ||
		pos.getCell(g.MineGrid) == CellMine ||
		pos.isOrHasNeighbour(s.Grid, CellMeT) ||
		pos.isOrHasNeighbour(s.Grid, CellMeNT)) &&
		!pos.sameAs(g.Me.Hq) {
		fmt.Fprintf(os.Stderr, "\t traversing (%d,%d)\n", pos.X, pos.Y)
		pos = pos.neighbour(pos.getIntCell(g.Op.DirGrid))
	}
	fmt.Fprintf(os.Stderr, "\t candidate at (%d,%d)\n", pos.X, pos.Y)
	if !pos.sameAs(g.Me.Hq) {
		fmt.Fprintf(os.Stderr, "\t accepted (%d,%d)\n", pos.X, pos.Y)
		return pos
	}
	fmt.Fprintf(os.Stderr, "\t rejected (%d,%d)\n", pos.X, pos.Y)
	return nil
}

func buildMinesAndTowers(s *State) {
	// build tower near HQ
	if (s.Op.ChainTrainWinNext || s.Op.MinUnitDistGoal <= 5) && s.Me.Gold > CostTower {
		pos := getHqTowerPosition()
		if pos.getCell(s.Grid) == CellMeA && pos.getCell(s.UnitGrid) == CellNeutral {
			fmt.Fprintf(os.Stderr, "%d: Build HQ tower\n", g.Turn)
			s.addBuildTower(pos)
		}
	}
	// build towers on Op ChainTrainWin path
	if (s.Op.ChainTrainWinNext || g.InTouch && s.Me.NbTowers < MaxTowersInTouch || s.NeutralPct < 0.2) &&
		s.Me.NbTowers < MaxTowers && s.Me.Gold > CostTower {
		if spot := findTowerSpotBeyondDist2(s, s.Op.MinDistGoal); spot != nil {
			s.addBuildTower(spot)
		} else {
			fmt.Fprintf(os.Stderr, "Couldn't find a tower spot beyond dist 2 starting at (%d,%d)\n", s.Op.MinDistGoal.X, s.Op.MinDistGoal.Y)
			if spot := findTowerSpotBeyondDist1(s, s.Op.MinDistGoal); spot != nil {
				s.addBuildTower(spot)
			} else {
				fmt.Fprintf(os.Stderr, "Couldn't find any tower spot starting at (%d,%d)\n", s.Op.MinDistGoal.X, s.Op.MinDistGoal.Y)
			}
		}
	}
	// build mine near HQ
	if s.Me.NbUnits >= Min1 &&
		s.Op.income() > s.Me.income() &&
		s.Me.NbMines == 0 &&
		s.Me.Gold > CostMine &&
		s.NeutralPct < 0.2 {
		pos := getHqMinePosition()
		if pos.getCell(s.Grid) == CellMeA && pos.getCell(s.UnitGrid) == CellNeutral {
			fmt.Fprintf(os.Stderr, "%d: Build HQ mine\n", g.Turn)
			s.addBuildMine(pos)
		}
	}
}

//---------------------------------------------------------------------------------------
//---------------------------------------------------------------------------------------

func main() {
	g.TurnTime = time.Now()
	initGame()
	var turnEval float64
	var eval float64
	for ; ; g.nextTurn() {
		s := &State{}
		s.init()

		// calculate chain train win cost / check before move
		s.Me.calculateChainTrainWin(true, true)
		s.Op.calculateChainTrainWin(true, false)
		s.evaluate("TURN START")
		fmt.Fprintf(os.Stderr, "%d: Full turn eval change: %.1f\n", g.Turn, s.Eval-turnEval)
		turnEval = s.Eval
		fmt.Fprintf(os.Stderr, "%d: OP MOVE eval change: %.1f\n", g.Turn, s.Eval-eval)
		eval = s.Eval

		// 0. look for BUILD MINE and/or TOWER commands
		buildMinesAndTowers(s)

		s.Me.calculateChainTrainWin(true, false)
		s.Op.calculateChainTrainWin(true, false)
		s.evaluate("AFTER BUILD")
		fmt.Fprintf(os.Stderr, "%d: BUILD eval change: %.1f\n", g.Turn, s.Eval-eval)
		eval = s.Eval

		// 1. look at MOVE commands
		moveUnits(s)

		s.Me.recalculateActiveArea()
		s.Op.recalculateActiveArea()

		// check chain train win after move
		s.Me.calculateChainTrainWin(false, true)
		s.Op.calculateChainTrainWin(true, false)
		s.evaluate("AFTER MOVE")
		fmt.Fprintf(os.Stderr, "%d: MOVE eval change: %.1f\n", g.Turn, s.Eval-eval)
		eval = s.Eval

		// 2. look at TRAIN commands
		candidateCmds := trainUnits(s)

		if candidateCmds == nil {
			fmt.Fprintf(os.Stderr, "%d: No TRAIN candidates\n", g.Turn)

		} else {
			fmt.Fprintf(os.Stderr, "%d: %d TRAIN candidates\n", g.Turn, len(candidateCmds.Candidates))
			for i, cmd := range candidateCmds.Candidates {
				if cmd.Level == 0 {
					//de-duped
					continue
				}
				cost := costTrain(cmd.Level)
				fmt.Fprintf(os.Stderr, "\t%d: TRAIN candidate: value %d, level %d at (%d,%d)\n", i, cmd.Value, cmd.Level, cmd.To.X, cmd.To.Y)
				fmt.Fprintf(os.Stderr, "\t%d: cost %d, gold %d, income %d, upkeep %d\n", i, cost, s.Me.Gold, s.Me.income(), s.Me.Upkeep)
				if cost <= s.Me.Gold && s.Me.income() >= s.Me.Upkeep {
					s.addTrain(cmd.To, cmd.Level)
					s.Me.recalculateActiveArea()
					s.Op.recalculateActiveArea()

					// check chain train win after each TRAIN cmd (as of next turn)
					s.Op.calculateChainTrainWin(true, false)
					s.Me.calculateChainTrainWin(true, false)
					s.evaluate("AFTER TRAIN")
					fmt.Fprintf(os.Stderr, "%d: TRAIN eval change: %.1f\n", g.Turn, s.Eval-eval)
					trainEvalChange := s.Eval - eval
					eval = s.Eval

					if AbortTrainCmdsOnNegativeEvalChg && trainEvalChange < TrainNegativeEvalPainTolerance {
						// abort train commands
						fmt.Fprintf(os.Stderr, "%d: Aborting last TRAIN command\n", g.Turn)
						s.Commands[len(s.Commands)-1].Type = CmdAbort
						break
					}
				} else {
					if DebugTrain && i < 10 {
						fmt.Fprintf(os.Stderr, "\tSkipping %d: value %d, level %d at (%d,%d)\n", i, cmd.Value, cmd.Level, cmd.To.X, cmd.To.Y)
					} else {
						fmt.Fprintf(os.Stderr, "\tSkipping %d candidates...\n", len(candidateCmds.Candidates)-i)
						break
					}
				}
			}
		}
		// fmt.Fprintln(os.Stderr, "Debug messages...")
		fmt.Println(s.action()) // Write action to stdout

		fmt.Fprintf(os.Stderr, "Turn %d. elapsed: %v, response: %v\n", g.Turn, time.Since(g.TurnTime), time.Since(g.RespTime))
		g.TurnTime = time.Now()
	} // for
}

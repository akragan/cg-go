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
	//algo
	AlgoNaive  = 0
	AlgoMinMax = 1

	// debug
	DebugChainTrainWin = false
	DebugActiveArea    = false
	DebugCapturable    = false
	DebugNeutral       = false
	DebugTrain         = false
	DebugBuildTower    = false
	DebugDistGrid      = false

	//options
	StandGroundL1 = true
	StandGroundL2 = true

	SortUnitsAsc = true

	MaxTowers = 1
	Min1      = 3

	MoveBackwards                   = false
	RandomDirsAtInitDistGrid        = false
	AbortTrainCmdsOnNegativeEvalChg = true
	TrainNegativeEvalPainTolerance  = -25.0
	NbEvaluatedTrainCandidates      = 50

	EvalDiscountRate    = 5.0
	EvalHqCaptureFactor = 100.0

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
	CellMeP  = 'P' // my tower-Protected cell
	CellMeN  = 'o' // my inactive cell
	CellMeH  = 'H' // my HQ
	CellMeHP = 'Q' // my tower-Protected HQ
	CellMeM  = 'M' // my active Mine
	CellMeMP = 'I' // my tower-Protected Mine
	CellMeMN = 'N' // my iNactive Mine
	CellMeT  = 'T' // my active Tower
	CellMeTN = 'F' // my iNactive Tower

	CellOpA  = 'X' // op active cell
	CellOpP  = 'p' // op tower-Protected cell
	CellOpN  = 'x' // op inactive cell
	CellOpH  = 'h' // op HQ
	CellOpHP = 'q' // op tower-Protected HQ
	CellOpM  = 'm' // op active mine
	CellOpMP = 'i' // op tower-Protected mine
	CellOpMN = 'n' // op iNactive Mine
	CellOpT  = 't' // op active Tower
	CellOpTN = 'f' // op iNactive Tower

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

func copyPosition(pos *Position) *Position {
	if pos != nil {
		return pos.copy()
	}
	return nil
}

func (pos *Position) copy() *Position {
	return &Position{
		X:    pos.X,
		Y:    pos.Y,
		Dist: pos.Dist,
	}
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

func (b *Building) copy() *Building {
	return &Building{
		Type:  b.Type,
		Owner: b.Owner,
		X:     b.X,
		Y:     b.Y,
	}
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

func copyUnit(u *Unit) *Unit {
	if u != nil {
		return u.copy()
	}
	return nil
}

func (u *Unit) copy() *Unit {
	return &Unit{
		Id:      u.Id,
		X:       u.X,
		Y:       u.Y,
		Owner:   u.Owner,
		Level:   u.Level,
		Freedom: u.Freedom,
	}
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
	Gold   int
	Income int
	Game   *GamePlayer
	State  *State
	Other  *Player

	NbUnits  int
	NbUnits1 int
	NbUnits2 int
	NbUnits3 int

	NbMines  int
	NbTowers int

	ActiveArea int
	Upkeep     int

	MinUnitDistGoal int       // distance to goal from the closest unit
	MinDistGoal     *Position // distance to goal from the closest active cell
	MinDistGoalUnit *Unit     // reference to unit if present on the closest active cell

	MinChainTrainWinCost    int
	ActualChainTrainWinCost int
	RoundsToHqCapture       float64
	MilitaryPower           int
	ExpectedMilitaryPower   int
}

func (p *Player) deepCopy() *Player {
	return &Player{
		Id:     p.Id,
		Gold:   p.Gold,
		Income: p.Income,
		Game:   p.Game,
		State:  nil,
		Other:  nil,

		NbUnits:  p.NbUnits,
		NbUnits1: p.NbUnits1,
		NbUnits2: p.NbUnits2,
		NbUnits3: p.NbUnits3,

		NbMines:  p.NbMines,
		NbTowers: p.NbTowers,

		ActiveArea: p.ActiveArea,
		Upkeep:     p.Upkeep,

		MinUnitDistGoal: p.MinUnitDistGoal,
		MinDistGoal:     copyPosition(p.MinDistGoal),
		MinDistGoalUnit: copyUnit(p.MinDistGoalUnit),

		MinChainTrainWinCost:    p.MinChainTrainWinCost,
		ActualChainTrainWinCost: p.ActualChainTrainWinCost,
		RoundsToHqCapture:       p.RoundsToHqCapture,
		MilitaryPower:           p.MilitaryPower,
		ExpectedMilitaryPower:   p.ExpectedMilitaryPower,
	}
}

func (s *State) addUnit(u *Unit) {
	p := s.player(u.Owner)
	p.NbUnits++
	switch u.Level {
	case 1:
		p.NbUnits1++
		p.Upkeep += CostKeep1
	case 2:
		p.NbUnits2++
		p.Upkeep += CostKeep2
	case 3:
		p.NbUnits3++
		p.Upkeep += CostKeep3
	}
}

func (p *Player) addActiveArea(pos *Position) {
	p.ActiveArea++
	var dist int
	if p.Game.Initialized {
		dist = pos.getIntCell(p.Game.DistGrid)
	} else {
		dist = 22
	}
	if dist < p.MinDistGoal.Dist {
		p.MinDistGoal.set(pos.X, pos.Y).Dist = dist
	}
	if DebugActiveArea {
		fmt.Fprintf(os.Stderr, "\t\t%s active area (%d) - added (%d,%d)\n", p.Game.Name, p.ActiveArea, pos.X, pos.Y)
	}
}

func (p *Player) activate(cell rune) rune {
	if p.Id == IdMe {
		return CellMeA
	}
	return CellOpA
}

func protect(cell rune) rune {
	switch cell {

	case CellMeA:
		return CellMeP
	case CellMeM:
		return CellMeMP
	case CellMeH:
		return CellMeHP

	case CellOpA:
		return CellOpP
	case CellOpM:
		return CellOpMP
	case CellOpH:
		return CellOpHP
	}
	return cell
}

func (s *State) protectNeighbours(pos *Position) {
	// set tower-protected cells
	for _, dir := range DirDRUL {
		nbrPos := pos.neighbour(dir)
		if nbrPos != nil {
			nbrPos.setCell(s.Grid, protect(nbrPos.getCell(s.Grid)))
		}
	} //end for
}

func (s *State) isProtectedIfActive(playerId int, pos *Position) bool {
	p := s.player(playerId)
	for _, dir := range DirDRUL {
		nbrPos := pos.neighbour(dir)
		if nbrPos != nil {
			cell := nbrPos.getCell(s.Grid)
			if p.isMyTower(cell) {
				return true
			}
		}
	}
	return false
}

func (this *Player) addActiveTower(bPos *Position) {
	s := this.State
	if this.Id == IdMe {
		bPos.setCell(s.Grid, CellMeT)
		s.protectNeighbours(bPos)
	} else {
		bPos.setCell(s.Grid, CellOpT)
		s.protectNeighbours(bPos)
	} //end if
}

func (p *Player) income() int {
	return p.ActiveArea + 4*p.NbMines - p.Upkeep
}

func (p *Player) expectedIncome() int {
	return p.ActiveArea + p.areaCapturableNextTurn() +
		4*p.NbMines - p.Upkeep
}

func (p *Player) myUnit(level int) rune {
	if p.Id == IdMe {
		switch level {
		case 1:
			return CellMeU
		case 2:
			return CellMeU2
		case 3:
			return CellMeU3
		}
	}
	switch level {
	case 1:
		return CellOpU
	case 2:
		return CellOpU2
	case 3:
		return CellOpU3
	}
	return CellNeutral
}

func (p *Player) isMyUnit(unitCell rune) bool {
	if p.Id == IdMe {
		return isMyUnitCell(unitCell)
	}
	return isOpUnitCell(unitCell)
}

func (p *Player) isEnemyUnit(unitCell rune) bool {
	if p.Id == IdMe {
		return isOpUnitCell(unitCell)
	}
	return isMyUnitCell(unitCell)
}

func (p *Player) isEnemyUnitLevel1(unitCell rune) bool {
	if p.Id == IdMe {
		return unitCell == CellOpU
	}
	return unitCell == CellMeU
}

func (p *Player) isEnemyUnitLevel2(unitCell rune) bool {
	if p.Id == IdMe {
		return unitCell == CellOpU2
	}
	return unitCell == CellMeU2
}

func (p *Player) isEnemyUnitLevel3(unitCell rune) bool {
	if p.Id == IdMe {
		return unitCell == CellOpU3
	}
	return unitCell == CellMeU3
}

func (p *Player) isEnemyUnitLevel2or3(unitCell rune) bool {
	if p.Id == IdMe {
		return unitCell == CellOpU2 || unitCell == CellOpU3
	}
	return unitCell == CellMeU2 || unitCell == CellMeU3
}

func (p *Player) isEnemyUnprotectedHQ(cell rune) bool {
	if p.Id == IdMe {
		return cell == CellOpH
	}
	return cell == CellMeH
}

func (p *Player) isEnemyProtectedHQ(cell rune) bool {
	if p.Id == IdMe {
		return cell == CellOpHP
	}
	return cell == CellMeHP
}

func (p *Player) isMyTower(cell rune) bool {
	if p.Id == IdMe {
		return cell == CellMeT
	}
	return cell == CellOpT
}

func (p *Player) isEnemyTower(cell rune) bool {
	if p.Id == IdMe {
		return cell == CellOpT
	}
	return cell == CellMeT
}

func (p *Player) isEnemyInactiveTower(cell rune) bool {
	if p.Id == IdMe {
		return cell == CellOpTN
	}
	return cell == CellMeTN
}

func (p *Player) isEnemyProtectedEmpty(cell rune) bool {
	if p.Id == IdMe {
		return cell == CellOpP
	}
	return cell == CellMeP
}

func (p *Player) isEnemyProtectedAny(cell rune) bool {
	return p.isEnemyTower(cell) ||
		p.isEnemyProtectedEmpty(cell) ||
		p.isEnemyProtectedMine(cell) ||
		p.isEnemyProtectedHQ(cell)
}

func (p *Player) myMine() rune {
	if p.Id == IdMe {
		return CellMeM
	}
	return CellOpM
}

func (p *Player) myTower() rune {
	if p.Id == IdMe {
		return CellMeT
	}
	return CellOpT
}

func (p *Player) myInactiveTower() rune {
	if p.Id == IdMe {
		return CellMeTN
	}
	return CellOpTN
}

func (p *Player) isEnemyUnprotectedMine(cell rune) bool {
	if p.Id == IdMe {
		return cell == CellOpM
	}
	return cell == CellMeM
}

func (p *Player) isEnemyProtectedMine(cell rune) bool {
	if p.Id == IdMe {
		return cell == CellOpMP
	}
	return cell == CellMeMP
}

func (p *Player) isEnemyInactiveMine(cell rune) bool {
	if p.Id == IdMe {
		return cell == CellOpMN
	}
	return cell == CellMeMN
}

func (p *Player) myEmptyActiveCell() rune {
	if p.Id == IdMe {
		return CellMeA
	}
	return CellOpA
}

func (p *Player) isMyEmptyActiveCell(cell rune) bool {
	if p.Id == IdMe {
		return cell == CellMeA
	}
	return cell == CellOpA
}

func (p *Player) isEnemyEmptyActiveCell(cell rune) bool {
	if p.Id == IdMe {
		return cell == CellOpA
	}
	return cell == CellMeA
}

func (p *Player) isEnemyEmptyInactiveCell(cell rune) bool {
	if p.Id == IdMe {
		return cell == CellOpN
	}
	return cell == CellMeN
}

func (p *Player) isMyActiveCell(cell rune) bool {
	if p.Id == IdMe {
		return isMyActiveCell(cell)
	}
	return isOpActiveCell(cell)
}

func (p *Player) isEnemyActiveCell(cell rune) bool {
	if p.Id == IdMe {
		return isOpActiveCell(cell)
	}
	return isMyActiveCell(cell)
}

func (s *State) calculateChainTrainWins(moveFirst bool, execute bool) bool {
	won := s.calculateChainTrainWin(IdMe, moveFirst, execute)
	s.calculateChainTrainWin(IdOp, true, false)
	return won
}

func (s *State) calculateChainTrainWin(playerId int, moveFirst bool, execute bool) bool {
	p := s.player(playerId)
	if DebugChainTrainWin {
		fmt.Fprintf(os.Stderr, "%d: [%s] Calculating ChainTrainWin: Gold=%d MinTrainChainCost=%d\n", g.Turn, p.Game.Name, p.Gold, p.MinDistGoal.Dist*CostTrain1)
	}
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
		if p.isEnemyProtectedAny(cell) || p.isEnemyInactiveTower(cell) || p.isEnemyUnitLevel2or3(unitCell) {
			level = 3
		}
		if moveFirst && isMyUnit && level == 1 { // fix to account for more free first moves of level 2 and 3
			// first move for free
			if DebugChainTrainWin {
				fmt.Fprintf(os.Stderr, "\t[%s] using free move first to move to (%d,%d) level=%d\n", p.Game.Name, pos.X, pos.Y, level)
			}
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
		if DebugChainTrainWin {
			fmt.Fprintf(os.Stderr, "\t[%s] Abort: Gold=%d ActualCost=%d\n", p.Game.Name, p.Gold, actualCost)
		}
		return false
	}
	if DebugChainTrainWin {
		fmt.Fprintf(os.Stderr, "\t[%s] Proceed: Gold=%d ActualCost=%d\n", p.Game.Name, p.Gold, actualCost)
	}
	if !execute {
		return false
	}
	for i, cmd := range cmds.Candidates {
		if cmd.Level == 0 { //move command
			s.addMove(playerId, cmd.Unit, cmd.From, cmd.To)
			if DebugChainTrainWin {
				fmt.Fprintf(os.Stderr, "\t%d: value %d, move %d to (%d,%d)\n", i, cmd.Value, cmd.Unit.Id, cmd.To.X, cmd.To.Y)
			}
		} else {
			s.addTrain(playerId, cmd.To, cmd.Level)
			if DebugChainTrainWin {
				fmt.Fprintf(os.Stderr, "\t%d: value %d, level %d at (%d,%d)\n", i, cmd.Value, cmd.Level, cmd.To.X, cmd.To.Y)
			}
		}
	}
	return true
}

func (p *Player) evaluate() {
	p.RoundsToHqCapture = 100.0
	if p.ActualChainTrainWinCost < p.Gold {
		p.RoundsToHqCapture = 0.0
	} else if p.expectedIncome() > 0 {
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
	Id          int
	Name        string
	Hq          *Position
	Other       *GamePlayer
	DistGrid    [][]int
	DirGrid     [][]int
	Initialized bool
}

func (s *State) calculateActiveAreas() {
	s.Me.recalculateActiveArea()
	s.Op.recalculateActiveArea()
}

func (p *Player) recalculateActiveArea() {
	if DebugActiveArea {
		fmt.Fprintf(os.Stderr, "\t\t%s recalculating active area (%d)\n", p.Game.Name, p.ActiveArea)
	}

	activeCells := make([][]rune, GridDim)
	for i := 0; i < GridDim; i++ {
		activeCells[i] = []rune(RowNeutral)
	}
	activeArea := 0
	pos := &Position{X: p.Game.Hq.X, Y: p.Game.Hq.Y, Dist: 0}
	todo := PositionQueue{pos}
	for !todo.IsEmpty() {
		todo, pos = todo.TakeFirst()
		activeCell := pos.getCell(activeCells)
		if activeCell != CellNeutral {
			continue
		}
		cell := pos.getCell(p.State.Grid)
		if p.isMyActiveCell(cell) {
			activeArea += 1
			pos.setCell(activeCells, CellMine)
			if DebugActiveArea {
				fmt.Fprintf(os.Stderr, "\t\t%s: %d active cells (%d,%d)\n", p.Game.Name, activeArea, pos.X, pos.Y)
			}
		} else {
			pos.setCell(activeCells, CellVoid)
		}
		dirs := DirDRUL
		for _, dir := range dirs {
			nbrPos := pos.neighbour(dir)
			if nbrPos != nil && p.isMyActiveCell(nbrPos.getCell(p.State.Grid)) {
				todo = todo.Put(nbrPos)
			} // if not visited
		} // for all dirs
	}
	activeAreaChg := activeArea - p.ActiveArea
	if activeAreaChg != 0 {
		fmt.Fprintf(os.Stderr, "\t\t%s: active area changed by %d (from %d to %d)\n", p.Game.Name, activeAreaChg, p.ActiveArea, activeArea)
		//p.ActiveArea = activeArea
		//TODO update active area
		//p.updateActive(activeCells)
	} else if DebugActiveArea {
		fmt.Fprintf(os.Stderr, "\t\t%s active area unchanged (%d)\n", p.Game.Name, p.ActiveArea)
	}
}

func (gp *GamePlayer) initDistGrid(grid [][]rune) {
	pos := &Position{X: gp.Other.Hq.X, Y: gp.Other.Hq.Y, Dist: 0}
	todo := PositionQueue{pos}
	for !todo.IsEmpty() {
		todo, pos = todo.TakeFirst()
		if pos.getIntCell(gp.DistGrid) != -1 {
			continue
		}
		//fmt.Fprintf(os.Stderr, "init DistGrid: (%d,%d):%d, queue size=%d\n", pos.X, pos.Y, pos.Dist, len(todo))
		if pos.getCell(grid) == CellVoid {
			pos.setIntCell(gp.DistGrid, InfDist)
		} else {
			pos.setIntCell(gp.DistGrid, pos.Dist)
			if pos.Dist != 0 {
				pos.setIntCell(gp.DirGrid, pos.findNeighbourDir(gp.DistGrid, pos.Dist-1))
			}
			dirs := DirDRUL
			if RandomDirsAtInitDistGrid {
				dirs = randDirs()
			}
			for _, dir := range dirs {
				nbrPos := pos.neighbour(dir)
				if nbrPos != nil && nbrPos.getIntCell(gp.DistGrid) == -1 {
					nbrPos.Dist = pos.Dist + 1
					todo = todo.Put(nbrPos)
					//fmt.Fprintf(os.Stderr, "\tdir=%v add (%d,%d):%d, queue size=%d\n", dir, nbrPos.X, nbrPos.Y, nbrPos.Dist, len(todo))
				} // if -1 (Dist not set)
			} // for all dirs
		} // if/else cell void
		//printDistGrid()
	} // for queue non-empty
	gp.Initialized = true
	if DebugDistGrid {
		printIntGrid(gp.Name+" DistGrid", gp.DistGrid)
		printIntGrid(gp.Name+" DirGrid", gp.DirGrid)
	}
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
	Algo           int
	Turn           int
	Eval           float64
	DiscountFactor float64

	TurnTime time.Time
	RespTime time.Time

	Me *GamePlayer
	Op *GamePlayer

	NbMines         int
	Mines           []*Position
	MineGrid        [][]rune
	InitNeutralArea int
	TotalArea       int
	InTouch         bool
}

func (g *Game) nextTurn() {
	g.Turn += 1
	g.DiscountFactor = math.Exp((float64(g.Turn)/100.0 - 1.0) * EvalDiscountRate)
}

func (g *Game) initGame() {
	g.Algo = AlgoNaive
	g.Turn = 0
	g.DiscountFactor = math.Exp(-1.0 * EvalDiscountRate)

	fmt.Scan(&g.NbMines)
	g.InitNeutralArea = 0
	g.TotalArea = 0
	g.InTouch = false
	g.Mines = make([]*Position, g.NbMines)
	g.MineGrid = make([][]rune, GridDim)

	g.Me = &GamePlayer{Id: IdMe, Name: "Me", Initialized: false}
	g.Me.DistGrid = make([][]int, GridDim)
	g.Me.DirGrid = make([][]int, GridDim)

	g.Op = &GamePlayer{Id: IdOp, Name: "Op", Initialized: false}
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

func (gp *GamePlayer) getHqTowerPosition() *Position {
	if gp.Hq.X == 0 {
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

func (gp *GamePlayer) getHqMinePosition() *Position {
	if gp.Hq.X == 0 {
		// build mine at (1,0)
		return &Position{X: 1, Y: 0}
	}
	// else build mine at (10,11)
	return &Position{X: 10, Y: 11}
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
	UnitGrid    [][]rune
	// state eval
	MilitaryPowerEval float64
	HqCaptureEval     float64
	Eval              float64
	// my commands to action
	Commands []*Command
}

func (s *State) deepCopy() *State {
	s2 := &State{
		Me:          s.Me.deepCopy(),
		Op:          s.Op.deepCopy(),
		Grid:        copyGrid(s.Grid),
		Neutral:     s.Neutral,
		NeutralPct:  s.NeutralPct,
		NbBuildings: s.NbBuildings,
		Buildings:   copyBuildings(s.Buildings),
		NbUnits:     s.NbUnits,
		Units:       copyUnits(s.Units),
		UnitGrid:    copyGrid(s.UnitGrid),

		MilitaryPowerEval: s.MilitaryPowerEval,
		HqCaptureEval:     s.HqCaptureEval,
		Eval:              s.Eval,

		Commands: []*Command{},
	}

	s2.Me.State = s2
	s2.Op.State = s2
	s2.Me.Other = s2.Op
	s2.Op.Other = s2.Me
	return s2
}

func (s *State) player(id int) *Player {
	if id == IdMe {
		return s.Me
	}
	return s.Op
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

func copyGrid(grid [][]rune) [][]rune {
	grid2 := make([][]rune, GridDim)
	for i := 0; i < GridDim; i++ {
		grid2[i] = make([]rune, GridDim)
		for j := 0; j < GridDim; j++ {
			grid2[i][j] = grid[i][j]
		}
	}
	return grid2
}

func copyBuildings(b []*Building) []*Building {
	n := len(b)
	b2 := make([]*Building, n)
	for i := 0; i < n; i++ {
		b2[i] = b[i].copy()
	}
	return b2
}

func copyUnits(u []*Unit) []*Unit {
	n := len(u)
	u2 := make([]*Unit, n)
	for i := 0; i < n; i++ {
		u2[i] = u[i].copy()
	}
	return u2
}

func (s *State) applyBuildings() {
	s.Me.NbMines = 0
	s.Op.NbMines = 0
	s.Me.NbTowers = 0
	s.Op.NbTowers = 0
	// reflect buildings on s.Grid
	// sort HQ/mines first, towers last - to protect correctly
	sort.Slice(s.Buildings, func(i, j int) bool { return s.Buildings[j].Type == TypeTower })
	for i := 0; i < s.NbBuildings; i++ {
		b := s.Buildings[i]
		bPos := b.Pos()
		bOwner := s.player(b.Owner)
		cell := bPos.getCell(s.Grid)

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
				if isMyActiveCell(cell) {
					bPos.setCell(s.Grid, CellMeM)
				} else {
					bPos.setCell(s.Grid, CellMeMN)
				}
				// TODO find out if inactive mines count towards building cost
				s.Me.NbMines++
			} else {
				if bPos.getCell(s.Grid) == CellOpA {
					bPos.setCell(s.Grid, CellOpM)
				} else {
					bPos.setCell(s.Grid, CellOpMN)
				}
				s.Op.NbMines++
			}
		case TypeTower:
			cell := bPos.getCell(s.Grid)
			if b.Owner == IdMe {
				// tower cell active or protected by another tower
				if isMyActiveCell(cell) {
					bOwner.addActiveTower(bPos)
				} else {
					bPos.setCell(s.Grid, CellMeTN)
				}
				s.Me.NbTowers++
			} else {
				// tower cell active or protected by another tower
				if isOpActiveCell(cell) {
					bOwner.addActiveTower(bPos)
				} else {
					bPos.setCell(s.Grid, CellOpTN)
				}
				s.Op.NbTowers++
			}
		}
	}
}

func (p *Player) isCapturable(level int, cell rune, unitCell rune) bool {
	switch level {
	case 1:
		return !p.isMyActiveCell(cell) &&
			!p.isEnemyProtectedAny(cell) &&
			!p.isEnemyInactiveTower(cell) &&
			!p.isEnemyUnit(unitCell)
	case 2:
		return !p.isMyActiveCell(cell) &&
			!p.isEnemyProtectedAny(cell) &&
			!p.isEnemyInactiveTower(cell) &&
			!p.isEnemyUnitLevel2or3(unitCell)
	case 3:
		return !p.isMyActiveCell(cell)
	}
	return false
}

func (p *Player) areaCapturableNextTurn() int {
	captured := make(map[int]bool)
	s := p.State
	for i := 0; i < s.NbUnits; i++ {
		u := s.Units[i]
		if u.Owner != p.Id {
			continue
		}
		pos := u.Pos()
		for _, dir := range DirDRUL {
			nbrPos := pos.neighbour(dir)
			if nbrPos != nil {
				nbrCell := nbrPos.getCell(s.Grid)
				nbrUnitCell := nbrPos.getCell(s.UnitGrid)
				if p.isCapturable(u.Level, nbrCell, nbrUnitCell) && !captured[nbrPos.toInt()] {
					captured[nbrPos.toInt()] = true
					break
				}
			}
		}
	} // for all units
	nbCapturable := len(captured)
	if DebugCapturable {
		fmt.Fprintf(os.Stderr, "%d: %s capturable next turn %d\n", g.Turn, p.Game.Name, nbCapturable)
	}
	return nbCapturable
}

func (s *State) applyUnits() {
	pos := &Position{}
	for i := 0; i < s.NbUnits; i++ {
		u := s.Units[i]
		p := s.player(u.Owner)
		pos.set(u.X, u.Y)
		if !g.InTouch && p.Id == IdMe && pos.findNeighbour(s.Grid, CellOpA) != -1 {
			g.InTouch = true
		}
		u.Freedom = pos.freedom(s.Grid)
		pos.setDistance(p.Game.Other.Hq)
		if p.MinUnitDistGoal > pos.Dist {
			p.MinUnitDistGoal = pos.Dist
		}
		if p.MinDistGoal.sameAs(pos) {
			p.MinDistGoalUnit = u
		}
		pos.setCell(s.UnitGrid, p.myUnit(u.Level))
		s.addUnit(u)
	} // for all units
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
				if g.Turn == 0 {
					g.TotalArea += 1
				}
				s.Me.addActiveArea(pos)
			} else if line[j] == CellOpA {
				if g.Turn == 0 {
					g.TotalArea += 1
				}
				s.Op.addActiveArea(pos)
			} else if line[j] == CellNeutral {
				if g.Turn == 0 {
					g.TotalArea += 1
					g.InitNeutralArea += 1
				}
				s.Neutral += 1
			}
		}
		s.UnitGrid[i] = []rune(RowNeutral)
	}
	s.NeutralPct = float32(s.Neutral) / float32(g.InitNeutralArea)
	s.Me.MinChainTrainWinCost = s.Me.MinDistGoal.Dist * CostTrain1
	s.Me.ActualChainTrainWinCost = s.Me.MinChainTrainWinCost
	s.Op.MinChainTrainWinCost = s.Op.MinDistGoal.Dist * CostTrain1
	s.Op.ActualChainTrainWinCost = s.Op.MinChainTrainWinCost

	if DebugNeutral {
		fmt.Fprintf(os.Stderr, "%d: NeutralPct=%v\n", g.Turn, s.NeutralPct)
	}
	// load buildings
	fmt.Scan(&s.NbBuildings)
	s.Buildings = make([]*Building, s.NbBuildings)
	for i := 0; i < s.NbBuildings; i++ {
		b := Building{}
		fmt.Scan(&b.Owner, &b.Type, &b.X, &b.Y)
		s.Buildings[i] = &b
	}
	s.applyBuildings()

	if g.Turn == 0 {
		g.Me.initDistGrid(s.Grid)
		g.Op.initDistGrid(s.Grid)
	}

	fmt.Scan(&s.NbUnits)
	s.Units = make([]*Unit, s.NbUnits)
	for i := 0; i < s.NbUnits; i++ {
		u := &Unit{}
		fmt.Scan(&u.Owner, &u.Id, &u.Level, &u.X, &u.Y)
		s.Units[i] = u
	}
	s.applyUnits()
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
	}
	s.Commands = []*Command{&Command{Type: CmdWait}}
}

func (s *State) addBuildMine(playerId int, at *Position) {
	if playerId == IdMe {
		s.Commands = append(s.Commands, &Command{Type: CmdBuildMine, X: at.X, Y: at.Y})
	}
	p := s.player(playerId)
	at.setCell(s.Grid, p.myMine())
	p.Gold -= CostMine
	p.NbMines += 1
}

func (s *State) addBuildTower(playerId int, at *Position) {
	if playerId == IdMe {
		s.Commands = append(s.Commands, &Command{Type: CmdBuildTower, X: at.X, Y: at.Y})
	}
	p := s.player(playerId)
	p.addActiveTower(at)
	p.Gold -= CostTower
	p.NbTowers += 1
}

func (s *State) addMove(playerId int, u *Unit, from *Position, to *Position) {
	if playerId == IdMe {
		s.Commands = append(s.Commands, &Command{Type: CmdMove, Info: u.Id, X: to.X, Y: to.Y})
	}
	p := s.player(playerId)
	if !from.sameAs(to) {
		to.setCell(s.UnitGrid, p.myUnit(u.Level))
		from.setCell(s.UnitGrid, CellNeutral)
		u.X = to.X
		u.Y = to.Y

		cell := to.getCell(s.Grid)
		if !p.isMyActiveCell(cell) {
			to.setCell(s.Grid, p.myEmptyActiveCell())
			p.addActiveArea(to)
		}
	}
}

func (s *State) addTrain(playerId int, at *Position, level int) {
	if playerId == IdMe {
		s.Commands = append(s.Commands, &Command{Type: CmdTrain, Info: level, X: at.X, Y: at.Y})
	}
	p := s.player(playerId)
	at.setCell(s.UnitGrid, p.myUnit(level))
	p.Gold -= costTrain(level)
	u := &Unit{Owner: playerId, Id: -1, Level: level, X: at.X, Y: at.Y}
	s.addUnit(u)
	s.Units = append(p.State.Units, u)
	s.NbUnits++
	cell := at.getCell(s.Grid)
	if DebugTrain {
		fmt.Fprintf(os.Stderr, "\t%s: training level %d at cell %s(%d,%d)\n", p.Game.Name, level, string(cell), at.X, at.Y)
	}
	if !p.isMyActiveCell(cell) {
		activeCell := p.activate(cell)
		at.setCell(s.Grid, activeCell)
		if s.isProtectedIfActive(playerId, at) {
			at.setCell(s.Grid, protect(activeCell))
		}
		p.addActiveArea(at)
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

func isMyUnitCell(cell rune) bool {
	return cell == CellMeU || cell == CellMeU2 || cell == CellMeU3
}

func isOpUnitCell(cell rune) bool {
	return cell == CellOpU || cell == CellOpU2 || cell == CellOpU3
}

func isAnyUnitCell(cell rune) bool {
	return isMyUnitCell(cell) || isOpUnitCell(cell)
}

func isMyActiveCell(cell rune) bool {
	return cell == CellMeA ||
		cell == CellMeP ||
		cell == CellMeH ||
		cell == CellMeHP ||
		cell == CellMeM ||
		cell == CellMeMP ||
		cell == CellMeT
}

func isOpActiveCell(cell rune) bool {
	return cell == CellOpA ||
		cell == CellOpP ||
		cell == CellOpH ||
		cell == CellOpHP ||
		cell == CellOpM ||
		cell == CellOpMP ||
		cell == CellOpT
}

func isMyInactiveCell(cell rune) bool {
	return cell == CellMeN ||
		cell == CellMeMN ||
		cell == CellMeTN
}

func isOpInactiveCell(cell rune) bool {
	return cell == CellOpN ||
		cell == CellOpMN ||
		cell == CellOpTN
}

func (p *Player) compactFactor(pos *Position, grid [][]rune) int {
	count := 0
	for _, dir := range DirDRUL {
		nbrPos := pos.neighbour(dir)
		if nbrPos != nil {
			nbrCell := nbrPos.getCell(grid)
			if p.isMyActiveCell(nbrCell) {
				count += 1
			}
		}
	}
	return count
}

func (p *Player) isWedge(pos *Position, grid [][]rune) bool {
	lPos := pos.neighbour(DirLeft)
	lOpA := lPos != nil && p.isEnemyActiveCell(lPos.getCell(grid))

	rPos := pos.neighbour(DirRight)
	rOpA := rPos != nil && p.isEnemyActiveCell(rPos.getCell(grid))

	uPos := pos.neighbour(DirUp)
	uOpA := uPos != nil && p.isEnemyActiveCell(uPos.getCell(grid))

	dPos := pos.neighbour(DirDown)
	dOpA := dPos != nil && p.isEnemyActiveCell(dPos.getCell(grid))

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

func (s *State) moveUnits(playerId int) {
	pos := &Position{}
	p := s.player(playerId)
	for i := 0; i < s.NbUnits; i++ {
		u := s.Units[i]
		if u.Owner != playerId || u.Id == -1 { // -1 for newly trained units that cannot move
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
			if p.isEnemyUnprotectedHQ(nbrCell) {
				candidateCmds.appendMove(u, pos, nbrPos, 20)
				continue
			}
			if u.Level == 3 && p.isEnemyProtectedHQ(nbrCell) {
				candidateCmds.appendMove(u, pos, nbrPos, 20)
				continue
			}
			// Op active TOWER capturing moves (only by l3 unit)
			if u.Level == 3 && p.isEnemyTower(nbrCell) && !p.isMyUnit(unitCell) {
				candidateCmds.appendMove(u, pos, nbrPos, 19)
				continue
			}
			// Op TOWER-protected mines and land capturing moves (only by l3 unit)
			if u.Level == 3 && p.isEnemyProtectedMine(nbrCell) && !p.isMyUnit(unitCell) {
				candidateCmds.appendMove(u, pos, nbrPos, 18)
				continue
			}
			if u.Level == 3 && p.isEnemyProtectedEmpty(nbrCell) && !p.isMyUnit(unitCell) {
				candidateCmds.appendMove(u, pos, nbrPos, 18)
				continue
			}
			// Op inactive TOWER capturing moves (only by l3 unit)
			if u.Level == 3 && p.isEnemyInactiveTower(nbrCell) && !p.isMyUnit(unitCell) {
				candidateCmds.appendMove(u, pos, nbrPos, 17)
				continue
			}
			// Op unit l3 capturing moves (only by l3 unit)
			if u.Level == 3 && p.isEnemyUnitLevel3(unitCell) && !p.isMyUnit(unitCell) {
				candidateCmds.appendMove(u, pos, nbrPos, 16)
				continue
			}
			// Op unit l2 capturing moves (only by l3 unit)
			if u.Level == 3 && p.isEnemyUnitLevel2(unitCell) && !p.isMyUnit(unitCell) {
				candidateCmds.appendMove(u, pos, nbrPos, 15)
				continue
			}
			// Op active MINE capturing moves (by any unit)
			if p.isEnemyUnprotectedMine(nbrCell) && !p.isMyUnit(unitCell) {
				candidateCmds.appendMove(u, pos, nbrPos, 14)
				continue
			}
			// Op unit l1 capturing moves (only by any l2 or l3 unit)
			if (u.Level == 3 || u.Level == 2) && p.isEnemyUnitLevel1(unitCell) &&
				!p.isEnemyProtectedAny(nbrCell) && !p.isMyUnit(unitCell) {
				candidateCmds.appendMove(u, pos, nbrPos, 13)
				//s.addMove(u, pos, nbrPos)
				continue
			}
			// Op INactive MINE capturing moves (by any unit)
			if p.isEnemyInactiveMine(nbrCell) && !p.isMyUnit(unitCell) {
				candidateCmds.appendMove(u, pos, nbrPos, 12)
				continue
			}
			// Op active land capturing moves (by any unit)
			// ++ priority for cells splitting Op territory
			// + priority for cells keeping my territory compact
			if p.isEnemyEmptyActiveCell(nbrCell) && !isAnyUnitCell(unitCell) {
				if s.Me.isWedge(nbrPos, s.Grid) {
					candidateCmds.appendMove(u, pos, nbrPos, 11)
				} else if s.Me.compactFactor(nbrPos, s.Grid) > 1 {
					candidateCmds.appendMove(u, pos, nbrPos, 10)
				} else {
					candidateCmds.appendMove(u, pos, nbrPos, 9)
				}
				continue
			}
			// Op INactive land capturing moves (by any unit)
			// + more priority for cells keeping my territory compact
			if p.isEnemyEmptyInactiveCell(nbrCell) && !p.isMyUnit(unitCell) {
				if s.Me.compactFactor(nbrPos, s.Grid) > 1 {
					candidateCmds.appendMove(u, pos, nbrPos, 8)
				} else {
					candidateCmds.appendMove(u, pos, nbrPos, 7)
				}
				continue
			}
			// new land capturing moves (by any unit)
			// + more priority for cells keeping my territory compact
			if nbrCell == CellNeutral && !p.isMyUnit(unitCell) {
				if s.Me.compactFactor(nbrPos, s.Grid) > 1 {
					candidateCmds.appendMove(u, pos, nbrPos, 5)
				} else {
					candidateCmds.appendMove(u, pos, nbrPos, 4)
				}
				continue
			}
			// standing my ground if faced with uncapturable enemy (lvl 1 and 2)
			// i.e. issuing invalid move command on purpose
			if StandGroundL2 && u.Level == 2 && p.isEnemyUnitLevel2(unitCell) ||
				StandGroundL1 && u.Level == 1 && p.isEnemyUnitLevel1(unitCell) {
				candidateCmds.appendMove(u, pos, pos, 0)
				continue
			}

			// just moving to another free cell (by any unit)
			// value depends on whether we're getting closer or further from Op Hq
			// 1 if closer, 0 if same, -1 if further
			if p.isMyEmptyActiveCell(nbrCell) && !p.isMyUnit(unitCell) {
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
			s.addMove(playerId, bestCmd.Unit, bestCmd.From, bestCmd.To)
		}
	}
}

//---------------------------------------------------------------------------------------
//---------------------------------------------------------------------------------------
// this produces dupe candidate train commands (in the same spots)
// as cells are neighbours of several other cells
// needs to be sorted and de-duped before execution
func (s *State) candidateTrainCmdsInNeighbourhood(playerId int, cmds *CommandSelector, pos *Position) {

	p := s.player(playerId)
	// 1. consider current cell (lowest value)
	cell := pos.getCell(s.Grid)
	unitCell := pos.getCell(s.UnitGrid)

	if p.isMyEmptyActiveCell(cell) && unitCell == CellNeutral {
		// copy pos
		pos := &Position{X: pos.X, Y: pos.Y}
		// consider level 1
		if p.Gold >= CostTrain1 {
			cmds.appendTrain(1, pos, 3-pos.getIntCell(p.Game.DistGrid))
		}
		// consider level 2
		if p.Gold >= CostTrain2 {
			cmds.appendTrain(2, pos, 1-pos.getIntCell(p.Game.DistGrid))
		}
		// consider level 3
		if p.Gold >= CostTrain3 {
			cmds.appendTrain(3, pos, 2-pos.getIntCell(p.Game.DistGrid))
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
		if p.isMyActiveCell(nbrCell) {
			// will be considered in its own right
			continue
		}
		nbrUnitCell := nbrPos.getCell(s.UnitGrid)
		bonus := 0
		if p.isWedge(nbrPos, s.Grid) {
			bonus += 3
		}
		if !nbrPos.isOrHasNeighbour(s.UnitGrid, p.myUnit(1)) {
			bonus += 1
		}

		if (nbrCell == CellNeutral || p.isEnemyEmptyInactiveCell(nbrCell)) && nbrUnitCell == CellNeutral {
			// consider level 1
			if p.Gold >= CostTrain1 {
				cmds.appendTrain(1, nbrPos, 6+bonus)
			}
			// consider level 2
			if p.Gold >= CostTrain2 {
				cmds.appendTrain(2, nbrPos, 4+bonus)
			}
			// consider level 3
			if p.Gold >= CostTrain3 {
				cmds.appendTrain(3, nbrPos, 5+bonus)
			}
		}

		if (p.isEnemyEmptyActiveCell(nbrCell) || p.isEnemyUnprotectedMine(nbrCell)) && nbrUnitCell == CellNeutral {
			// consider level 1
			if p.Gold >= CostTrain1 {
				cmds.appendTrain(1, nbrPos, 9+bonus)
			}
			// consider level 2
			if p.Gold >= CostTrain2 {
				cmds.appendTrain(2, nbrPos, 8+bonus)
			}
			// consider level 3
			if p.Gold >= CostTrain3 {
				cmds.appendTrain(3, nbrPos, 7+bonus)
			}
		}

		if p.isEnemyUnitLevel1(nbrUnitCell) && !p.isEnemyProtectedEmpty(nbrCell) {
			// consider level 2 and 3
			if p.Gold >= CostTrain2 {
				cmds.appendTrain(2, nbrPos, 11+bonus)
			}
			// consider level 3
			if p.Gold >= CostTrain3 {
				cmds.appendTrain(3, nbrPos, 10+bonus)
			}
		}

		if p.isEnemyUnitLevel2(nbrUnitCell) {
			// consider level 3
			if p.Gold >= CostTrain3 {
				cmds.appendTrain(3, nbrPos, 12+bonus)
			}
		}

		if p.isEnemyProtectedAny(nbrCell) {
			// consider level 3
			if p.Gold >= CostTrain3 {
				cmds.appendTrain(3, nbrPos, 13+bonus)
			}
		}

		if p.isEnemyUnitLevel3(nbrUnitCell) {
			// consider level 3
			if p.Gold >= CostTrain3 {
				cmds.appendTrain(3, nbrPos, 15+bonus)
			}
		}

		if p.isEnemyProtectedHQ(nbrCell) {
			// consider level 3
			if p.Gold >= CostTrain3 {
				cmds.appendTrain(3, nbrPos, 100)
			}
		}

		if p.isEnemyUnprotectedHQ(nbrCell) {
			// consider level 1
			if p.Gold >= CostTrain1 {
				cmds.appendTrain(1, nbrPos, 100)
			}
		}

	} //for dir
}

func (s *State) trainUnits(playerId int) *State {
	p := s.player(playerId)
	if p.Gold < CostTrain1 {
		// no gold to train any units
		return s
	}
	pos := &Position{}
	candidateCmds := &CommandSelector{}
	for j := 0; j < GridDim; j++ {
		for i := 0; i < GridDim; i++ {
			pos.set(i, j)
			cell := pos.getCell(s.Grid)
			if !p.isMyActiveCell(cell) {
				// can only train on and next to active area
				continue
			}
			s.candidateTrainCmdsInNeighbourhood(playerId, candidateCmds, pos)
		} // for i
	} // for j

	// sort, dedupe and execute
	candidateCmds.sort()
	candidateCmds.dedupe()

	if candidateCmds == nil {
		fmt.Fprintf(os.Stderr, "%d: No TRAIN candidates\n", g.Turn)

	} else {
		fmt.Fprintf(os.Stderr, "%d: %d TRAIN candidates\n", g.Turn, len(candidateCmds.Candidates))
		for i, cmd := range candidateCmds.Candidates {
			if cmd.Level == 0 {
				//de-duped
				continue
			}
			p = s.player(playerId)
			cost := costTrain(cmd.Level)
			fmt.Fprintf(os.Stderr, "%d: %dth TRAIN candidate: value %d, level %d at (%d,%d)\n", g.Turn, i, cmd.Value, cmd.Level, cmd.To.X, cmd.To.Y)
			fmt.Fprintf(os.Stderr, "\t%d: cost %d, gold %d, income %d, upkeep %d\n", i, cost, p.Gold, p.income(), p.Upkeep)
			if i < NbEvaluatedTrainCandidates && cost <= p.Gold && p.income() >= p.Upkeep {
				eval := s.Eval
				s2 := s.deepCopy()
				p2 := s2.player(playerId)
				s2.addTrain(p2.Id, cmd.To, cmd.Level)
				s2.moveUnits(p2.Other.Id)
				// evaluate after each TRAIN cmd (and opponent move)
				s2.calculateActiveAreas()
				s2.calculateChainTrainWins(true, false)
				s2.evaluate("AFTER TRAIN and opponent move")

				fmt.Fprintf(os.Stderr, "\t%d: TRAIN eval change: %.1f\n", g.Turn, s2.Eval-eval)
				if s2.Eval >= eval {
					eval = s2.Eval
					fmt.Fprintf(os.Stderr, "\t%s: appending TRAIN command, my gold=%d\n", p2.Game.Name, p2.Gold)
					if p2.Id == IdMe {
						s2.Commands = append(s.Commands, s2.Commands...)
					}
					s = s2
				} else {
					fmt.Fprintf(os.Stderr, "\tskipping TRAIN command\n")
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
	return s
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

func (s *State) findTowerSpotBeyondDist2(playerId int, pos *Position) *Position {
	p := s.player(playerId)
	for ; !pos.sameAs(p.Game.Hq); pos = pos.neighbour(pos.getIntCell(p.Game.Other.DirGrid)) {
		cell := pos.getCell(s.Grid)
		unitCell := pos.getCell(s.UnitGrid)
		mineCell := pos.getCell(g.MineGrid)
		if p.isMyEmptyActiveCell(cell) &&
			unitCell == CellNeutral &&
			mineCell != CellMine &&
			!pos.isOrHasNeighbourAtDist2(s.Grid, p.myTower()) &&
			!pos.isOrHasNeighbourAtDist2(s.Grid, p.myInactiveTower()) {
			break
		}
		if DebugBuildTower {
			fmt.Fprintf(os.Stderr, "\t find tower dist2: traversing (%d,%d)\n", pos.X, pos.Y)
		}
	}
	if !pos.sameAs(p.Game.Hq) {
		if DebugBuildTower {
			fmt.Fprintf(os.Stderr, "%d: Tower candidate at (%d,%d)\n", g.Turn, pos.X, pos.Y)
		}
		return pos
	}
	return nil
}

func (s *State) findTowerSpotBeyondDist1(playerId int, pos *Position) *Position {
	p := s.player(playerId)
	for ; !pos.sameAs(p.Game.Hq); pos = pos.neighbour(pos.getIntCell(p.Game.Other.DirGrid)) {
		cell := pos.getCell(s.Grid)
		unitCell := pos.getCell(s.UnitGrid)
		mineCell := pos.getCell(g.MineGrid)
		if p.isMyEmptyActiveCell(cell) &&
			unitCell == CellNeutral &&
			mineCell != CellMine &&
			!pos.isOrHasNeighbour(s.Grid, p.myTower()) &&
			!pos.isOrHasNeighbour(s.Grid, p.myInactiveTower()) {
			break
		}
		if DebugBuildTower {
			fmt.Fprintf(os.Stderr, "\t find tower dist1: traversing (%d,%d)\n", pos.X, pos.Y)
		}
	}
	if !pos.sameAs(p.Game.Hq) {
		if DebugBuildTower {
			fmt.Fprintf(os.Stderr, "%d: Tower candidate at (%d,%d)\n", pos.X, pos.Y)
		}
		return pos
	}
	return nil
}

func (s *State) buildMinesAndTowers(playerId int) {
	p := s.player(playerId)
	if p.NbTowers < MaxTowers && p.Gold > CostTower {
		// build tower near HQ
		spot := p.Game.getHqTowerPosition()
		cell := spot.getCell(s.Grid)
		unitCell := spot.getCell(s.UnitGrid)
		if p.isMyEmptyActiveCell(cell) && unitCell == CellNeutral {
			fmt.Fprintf(os.Stderr, "%d: Build HQ tower\n", g.Turn)
			s.addBuildTower(playerId, spot)
		} else if spot = s.findTowerSpotBeyondDist2(playerId, p.Other.MinDistGoal); spot != nil {
			// build towers on Op ChainTrainWin path
			s.addBuildTower(playerId, spot)
		} else {
			if DebugBuildTower {
				fmt.Fprintf(os.Stderr, "\tCouldn't find a tower spot beyond dist 2 starting at (%d,%d)\n", p.Other.MinDistGoal.X, p.Other.MinDistGoal.Y)
			}
			if spot = s.findTowerSpotBeyondDist1(playerId, p.Other.MinDistGoal); spot != nil {
				s.addBuildTower(playerId, spot)
			} else if DebugBuildTower {
				fmt.Fprintf(os.Stderr, "\tCouldn't find any tower spot starting at (%d,%d)\n", p.Other.MinDistGoal.X, p.Other.MinDistGoal.Y)
			}
		}
	}
	// build mine near HQ
	if p.NbMines == 0 && p.Gold > CostMine {
		spot := p.Game.getHqMinePosition()
		cell := spot.getCell(s.Grid)
		unitCell := spot.getCell(s.UnitGrid)
		if p.isMyEmptyActiveCell(cell) && unitCell == CellNeutral {
			fmt.Fprintf(os.Stderr, "%d: Build HQ mine\n", g.Turn)
			s.addBuildMine(playerId, spot)
		}
	}
}

//---------------------------------------------------------------------------------------
//---------------------------------------------------------------------------------------

func naiveAlgo(s *State) *State {
	eval := s.Eval
	s2 := s.deepCopy()
	s2.moveUnits(IdOp)
	s2.calculateChainTrainWins(true, false)
	s2.evaluate("DO NOTHING scenario")
	fmt.Fprintf(os.Stderr, "%d: DO NOTHING eval change: %.1f\n", g.Turn, s2.Eval-eval)
	eval = s2.Eval

	// 0. look for BUILD MINE and/or TOWER commands
	s2 = s.deepCopy()
	s2.buildMinesAndTowers(IdMe)
	s2.moveUnits(IdOp)
	// evaluate after BUILD cmds - no change to active areas
	if len(s2.Commands) > 0 {
		s2.calculateChainTrainWins(true, false)
		s2.evaluate("AFTER BUILD")
		fmt.Fprintf(os.Stderr, "%d: BUILD eval change: %.1f\n", g.Turn, s2.Eval-eval)
		if s2.Eval >= eval {
			eval = s2.Eval
			fmt.Fprintf(os.Stderr, "\tappending BUILD commands\n")
			s2.Commands = append(s.Commands, s2.Commands...)
			s = s2
		} else {
			fmt.Fprintf(os.Stderr, "\tskipping BUILD commands\n")
		}
	}
	// 1. look at MOVE commands
	s2 = s.deepCopy()
	s2.moveUnits(IdMe)
	s2.moveUnits(IdOp)
	// evaluate after move cmds
	if len(s2.Commands) > 0 {
		s2.calculateActiveAreas()
	}
	won := s2.calculateChainTrainWins(false, true)
	if won {
		s2.Commands = append(s.Commands, s2.Commands...)
		s = s2
	} else {
		if len(s2.Commands) > 0 {
			s2.evaluate("AFTER MOVE")
			fmt.Fprintf(os.Stderr, "%d: MOVE eval change: %.1f\n", g.Turn, s2.Eval-eval)
			if s2.Eval >= eval {
				eval = s2.Eval
				fmt.Fprintf(os.Stderr, "\tappending MOVE commands\n")
				s2.Commands = append(s.Commands, s2.Commands...)
				s = s2
			} else {
				fmt.Fprintf(os.Stderr, "\tskipping MOVE commands\n")
			}
		}
		// 2. look at TRAIN commands
		s = s.trainUnits(IdMe)
	}
	return s
}

func minMaxAlgo(s *State) *State {
	return s
}

func main() {
	g.TurnTime = time.Now()
	g.initGame()
	for ; ; g.nextTurn() {
		s := &State{}
		s.init()
		g.RespTime = time.Now()

		// check forced win on new turn before any scenarios
		won := s.calculateChainTrainWins(true, true)
		if !won {
			s.evaluate("NEW TURN")
			fmt.Fprintf(os.Stderr, "%d: Full turn eval change: %.1f\n", g.Turn, s.Eval-g.Eval)
			g.Eval = s.Eval

			switch g.Algo {
			case AlgoNaive:
				s = naiveAlgo(s)
			case AlgoMinMax:
				s = minMaxAlgo(s)
			}
		}
		fmt.Println(s.action()) // Write action to stdout

		fmt.Fprintf(os.Stderr, "Turn %d. elapsed: %v, response: %v\n", g.Turn, time.Since(g.TurnTime), time.Since(g.RespTime))
		g.TurnTime = time.Now()
	} // for
}

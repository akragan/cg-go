package main

import "fmt"
import "sort"
import "os"
//import "bufio"
//import "strings"

const(
    GridDim = 12

    IdMe = 0
    IdOp = 1
    IdVoid = -1
    
    CmdWait = 0
    CmdMove = 1
    CmdTrain = 2
    CmdBuildMine = 3
    CmdBuildTower = 4
    
    TypeHq = 0
    TypeMine = 1
    TypeTower = 2
    
    CostTrain1 = 10
    CostTrain2 = 20
    CostTrain3 = 30

    CostKeep1 = 1
    CostKeep2 = 4
    CostKeep3 = 20
    
    CostMine1 = 20
    
    CellVoid = '#'
    CellNeutral = '.'
    CellMine = '$'

    CellMeA = 'O'
    CellMeNA = 'o'
    CellMeH = 'H'
    CellMeM = 'M'

    CellOpA = 'X'
    CellOpNA = 'x'
    CellOpH = 'h'
    CellOpM = 'm'

    CellMeU = 'U'
    CellMeU2 = 'K'
    CellMeU3 = 'G'

    CellOpU = 'u'
    CellOpU2 = 'k'
    CellOpU3 = 'g'

    DirLeft = 0
    DirUp = 1
    DirRight = 2
    DirDown = 3
    
    Min1 = 3
    Min2 = 2
)

var(
    g = &Game{}
    
    DirDRUL = []int{DirDown, DirRight, DirUp, DirLeft}
    DirLURD = []int{DirLeft, DirUp, DirRight, DirDown}
)

func distance(x1 int, y1 int, x2 int, y2 int) int{
    dist:= 0
    if x1>x2 {
        dist+= x1-x2
    }else{
        dist+= x2-x1
    }
    if y1>y2 {
        dist+= y1-y2
    }else{
        dist+= y2-y1
    }
    return dist
}

type HasPosition interface{
    Pos() *Position
}

type Position struct{
    X int
    Y int
    Dist int
}

func (this *Position) Pos() *Position{
    return this
}

func (this *Position) setDistance(other *Position) int{
    this.Dist= distance(this.X, this.Y, other.X, other.Y)
    return this.Dist
}

// unsafe
func (this *Position) getCell(grid [][]rune) rune {
    return grid[this.Y][this.X]
}

func (this *Position) setCell(grid [][]rune, cell rune) {
    grid[this.Y][this.X]= cell
}

func (this *Position) set(x int, y int){
    this.X= x
    this.Y= y
}

func (this *Position) neighbour(direction int) *Position{
    var n *Position
    switch direction{
        case DirLeft: if this.X>0 {
            n= &Position{X:this.X-1, Y:this.Y}
        }
        case DirRight: if this.X<GridDim-1 {
            n= &Position{X:this.X+1, Y:this.Y}
        }
        case DirUp: if this.Y>0 {
            n= &Position{X:this.X, Y:this.Y-1}
        }
        case DirDown: if this.Y<GridDim-1 {
            n= &Position{X:this.X, Y:this.Y+1}
        }
    }
    return n
}


type Player struct{
    Id int
    Gold int
    Income int
    
    NbUnits int
    NbUnits1 int
    NbUnits2 int
    NbUnits3 int
    
    NbMines int
    MineSpots []*Position
    
    ActiveArea int
    Upkeep int
}

func (this *Player) addUnit(u *Unit){
    this.NbUnits++
    switch u.Level{
        case 1: 
            this.NbUnits1++
            this.Upkeep+= CostKeep1
        case 2: 
            this.NbUnits2++
            this.Upkeep+= CostKeep2
        case 3: 
            this.NbUnits3++
            this.Upkeep+= CostKeep3
    }
}

func (this *Player) income() int{
    return this.ActiveArea + 4*this.NbMines - this.Upkeep
}

func (this *Player) mineCost() int{
    return CostMine1 + this.NbMines*4
}

type Building struct {
    Type int
    Owner int
    X int
    Y int
}

func (this *Building) Pos() *Position{
    return &Position{X:this.X, Y:this.Y}
}

type Unit struct {
    Id int
    X int
    Y int
    Owner int
    Level int
}

func (this *Unit) Pos() *Position{
    return &Position{X:this.X, Y:this.Y}
}

type Command struct {
    Type int
    Info int
    X int
    Y int
}

func (this *Command) Pos() *Position{
    return &Position{X:this.X, Y:this.Y}
}


type Game struct{
    NbMines int
    Mines []*Position
    MineGrid [][]rune
    HqMe *Position
    HqOp *Position
}

func (game *Game) init(){
    fmt.Scan(&g.NbMines)
    g.Mines= make([]*Position, g.NbMines)
    g.MineGrid= make([][]rune, GridDim)
    for i:=0; i<GridDim; i++{
        g.MineGrid[i]= make([]rune, GridDim)
    }
    for i:=0; i<g.NbMines; i++ {
        mine := &Position{}
        fmt.Scan(&mine.X, &mine.Y)
        g.Mines[i]= mine
        mine.setCell(g.MineGrid, CellMine)
    }    
}

type State struct {
    Me *Player
    Op *Player
    Grid [][]rune
    NbBuildings int
    Buildings []*Building
    NbUnits int
    Units []*Unit
    UnitGrid [][]rune

    Commands []*Command
}

func (s *State) init(){
    
    s.Me= &Player{}
    fmt.Scan(&s.Me.Gold)
    fmt.Scan(&s.Me.Income)

    s.Op= &Player{}    
    fmt.Scan(&s.Op.Gold)
    fmt.Scan(&s.Op.Income)
    
    s.Grid= make([][]rune, GridDim)
    s.UnitGrid= make([][]rune, GridDim)
    for i := 0; i < GridDim; i++ {
        var line string
        fmt.Scan(&line)
        //fmt.Fprintf(os.Stderr, "%v\n", line)
        s.Grid[i]= []rune(line)
        for j:=0; j<GridDim; j++ {
            if line[j]==CellMeA{
                s.Me.ActiveArea++  
            }else if line[j]==CellOpA{
                s.Op.ActiveArea++
            }
        }
        s.UnitGrid[i]= []rune("............")
    }
    
    fmt.Scan(&s.NbBuildings)
    s.Buildings= make([]*Building, s.NbBuildings)
    for i := 0; i < s.NbBuildings; i++ {
        b := Building{}
        fmt.Scan(&b.Owner, &b.Type, &b.X, &b.Y)
        s.Buildings[i]= &b
        bPos:= b.Pos()
        switch b.Type{
            case TypeHq:
                if b.Owner == IdMe{
                    g.HqMe= bPos
                    bPos.setCell(s.Grid, CellMeH)
                }else{
                    g.HqOp= bPos
                    bPos.setCell(s.Grid, CellOpH)
                }
            case TypeMine:
                if b.Owner == IdMe{
                    // TODO check is mine is on active cell before overriding
                    bPos.setCell(s.Grid, CellMeM)
                    s.Me.NbMines++
                }else{
                    bPos.setCell(s.Grid, CellOpM)
                    s.Op.NbMines++
                }
        }
    }
    
    fmt.Scan(&s.NbUnits)
    s.Units= make([]*Unit, s.NbUnits)
    pos:= &Position{}
    for i := 0; i < s.NbUnits; i++ {
        u := &Unit{}
        fmt.Scan(&u.Owner, &u.Id, &u.Level, &u.X, &u.Y)
        s.Units[i]= u
        pos.set(u.X, u.Y)
        if u.Owner==IdMe{
            switch u.Level{
            case 1:
                pos.setCell(s.UnitGrid, CellMeU)
            case 2:
                pos.setCell(s.UnitGrid, CellMeU2)
            case 3:
                pos.setCell(s.UnitGrid, CellMeU3)
            }
            s.Me.addUnit(u)
        }else{
            switch u.Level{
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

    s.Commands= []*Command{&Command{Type:CmdWait}}
}

func (s *State) addMove(u *Unit, from *Position, to *Position){
    s.Commands= append(s.Commands, &Command{Type:CmdMove, Info:u.Id, X:to.X, Y:to.Y})
    to.setCell(s.Grid, CellMeA)
    to.setCell(s.UnitGrid, CellMeU)
    from.setCell(s.UnitGrid, CellNeutral)
}

func (s *State) addTrain(at *Position, level int){
    s.Commands= append(s.Commands, &Command{Type:CmdTrain, Info:level, X:at.X, Y:at.Y})
    at.setCell(s.UnitGrid, CellMeU)
    switch level{
        case 1: 
            s.Me.Gold-= CostTrain1
        case 2: 
            s.Me.Gold-= CostTrain2
        case 3: 
            s.Me.Gold-= CostTrain3
    }
}

func (s *State) action() string{
    cmdsStr:= ""
    for i:=0; i<len(s.Commands); i++ {
        if i>0 {
            cmdsStr+= ";"
        }
        cmd:= s.Commands[i]
        switch cmd.Type{
        case CmdWait: 
            cmdsStr+= "WAIT" 
        case CmdTrain:
            cmdsStr+= fmt.Sprintf("TRAIN %d %d %d", cmd.Info, cmd.X, cmd.Y)
        case CmdMove:
            cmdsStr+= fmt.Sprintf("MOVE %d %d %d", cmd.Info, cmd.X, cmd.Y)
        case CmdBuildMine:
            cmdsStr+= fmt.Sprintf("BUILD MINE %d %d", cmd.X, cmd.Y)
        case CmdBuildTower:
            cmdsStr+= fmt.Sprintf("BUILD TOWER %d %d", cmd.X, cmd.Y)
        }
    }
    cmdsStr+= fmt.Sprintf(";MSG A:%d U:%d I:%d", s.Me.ActiveArea, s.Me.NbUnits, s.Me.income())
    return cmdsStr
}


type CommandSelector struct {
    Candidates []*CandidateCommand
}

type CandidateCommand struct {
    Unit *Unit
    From *Position
    To *Position
    Value int
}

func (this *CommandSelector) appendMove(u *Unit, from *Position, to *Position, value int) {
    this.Candidates= append(this.Candidates, &CandidateCommand{
        Unit: u,
        From: from,
        To: to,
        Value: value,
    })
}

func (this *CommandSelector) bestCommand() *CandidateCommand{
    if len(this.Candidates)==0 {
        return nil
    }
    sort.Slice(this.Candidates, func(i, j int) bool { return this.Candidates[i].Value > this.Candidates[j].Value })
    return this.Candidates[0]
}

func myUnitCell(cell rune) bool{
    return cell == CellMeU || cell == CellMeU2 || cell == CellMeU3    
}

func moveUnits(s *State){
    pos:= &Position{}
    dirs:= DirDRUL
    if g.HqMe.X != 0{
        dirs= DirLURD
    }
    for i:=0; i < s.NbUnits; i++ {
        u:= s.Units[i]
        if u.Owner != IdMe {
            continue
        }
        pos.set(u.X, u.Y)
        //fmt.Fprintf(os.Stderr, "Unit: %d Pos: %d %d HQ: %d %d \n", u.Id, pos.X, pos.Y, g.HqMe.X, g.HqMe.Y)
        candidateCmds:= CommandSelector{}
        for _, dir := range(dirs) {
            nbrPos:= pos.neighbour(dir)
            if nbrPos != nil{

                nbrCell:= nbrPos.getCell(s.Grid)
                unitCell:= nbrPos.getCell(s.UnitGrid)

                if nbrCell == CellVoid {
                    continue
                }
                // value 10 == HQ capturing moves
                if nbrCell == CellOpH {
                    candidateCmds.appendMove(u, pos, nbrPos, 10)
                    continue
                }
                // value 8 == unit l3 capturing moves
                if u.Level==3 && unitCell == CellOpU3 && !myUnitCell(unitCell){
                    candidateCmds.appendMove(u, pos, nbrPos, 8)
                    continue
                }
                // value 7 == unit l2 capturing moves
                if u.Level==3 && unitCell == CellOpU2 && !myUnitCell(unitCell){
                    candidateCmds.appendMove(u, pos, nbrPos, 7)
                    continue
                }
                // value 6 == active MINE capturing moves
                if nbrCell == CellOpM && !myUnitCell(unitCell) {
                    candidateCmds.appendMove(u, pos, nbrPos, 6)
                    continue
                }
                // value 5 = unit l1 capturing moves
                if (u.Level==3 || u.Level==2) && unitCell == CellOpU && !myUnitCell(unitCell){
                    candidateCmds.appendMove(u, pos, nbrPos, 5)
                    //s.addMove(u, pos, nbrPos)
                    continue
                }
                // TODO value 4 == INactive MINE capturing moves
                //if nbrCell == CellOpM && !myUnitCell(unitCell) {
                //    candidateCmds.appendMove(u, pos, nbrPos, 4)
                //    continue
                //}
                // value 3 = active opponent land capturing moves
                if nbrCell == CellOpA && !myUnitCell(unitCell) {
                    candidateCmds.appendMove(u, pos, nbrPos, 3)
                    continue
                }
                // value 2 = INactive opponent land capturing moves
                if nbrCell == CellOpNA && !myUnitCell(unitCell) {
                    candidateCmds.appendMove(u, pos, nbrPos, 2)
                    continue
                }
                // value 1 = new land capturing moves
                if nbrCell == CellNeutral && !myUnitCell(unitCell) {
                    candidateCmds.appendMove(u, pos, nbrPos, 1)
                    continue
                }
                // value 0 = just moving to another free cell
                if nbrCell == CellMeA && !myUnitCell(unitCell) {
                    candidateCmds.appendMove(u, pos, nbrPos, 0)
                    continue
                }
            }
        }//for dir 
        // pick the best move for unit 
        if bestCmd:= candidateCmds.bestCommand(); bestCmd!=nil{
            fmt.Fprintf(os.Stderr, "Unit:%d, Candidates:%d, Best:%d X:%d Y:%d\n", bestCmd.Unit.Id, len(candidateCmds.Candidates), bestCmd.Value, bestCmd.To.X, bestCmd.To.Y)
            s.addMove(bestCmd.Unit, bestCmd.From, bestCmd.To)
        }
    }
}

func trainUnitInNeighbourhood(s *State, pos *Position, dirs []int) bool{
    for _, dir:= range(dirs) {
        nbrPos:= pos.neighbour(dir)
        if nbrPos != nil{
            nbrCell:= nbrPos.getCell(s.Grid)
            unitCell:= nbrPos.getCell(s.UnitGrid)
            if nbrCell == CellNeutral && unitCell == CellNeutral {
                if s.Me.NbUnits<Min1 || s.Me.NbUnits<int(1.25*float32(s.Op.NbUnits1)) &&
                    s.Me.Gold > CostTrain1 {
                    s.addTrain(nbrPos, 1)
                }else if s.Me.NbUnits2>Min2 && 
                        s.Me.NbUnits2>s.Op.NbUnits2 && 
                        s.Me.income() > 2*CostKeep3 &&
                        s.Me.Gold > CostTrain3{
                    s.addTrain(nbrPos, 3)
                }else if s.Me.income() > 2*CostKeep2 &&
                        s.Me.Gold > CostTrain2{
                    s.addTrain(nbrPos, 2)
                }
                //TRAIN only one
                return true
            }    
        }
    }//for dir     
    return false
}

func trainUnits(s *State){
    pos:= &Position{}
    dirs:= DirLURD
    if g.HqMe.X != 0{
        dirs= DirDRUL
    }     
    
    // train in new areas
    for j:= 0; j < GridDim; j++ {
        for i:= 0; i < GridDim; i++ {
            if s.Me.Gold < CostTrain1 {
                // no gold to train any units
                return
            }
            pos.set(i, j)
            cell:= pos.getCell(s.Grid)
            if cell!=CellMeA && cell!=CellMeH && cell!=CellMeM {
                continue
            }
            //TODO train more than one
            if trainUnitInNeighbourhood(s, pos, dirs){
                return
            }
        }// for i
    }// for j
    
    // no new are found, train at headquarters
    pos.set(g.HqMe.X, g.HqMe.Y)
    trainUnitInNeighbourhood(s, pos, dirs)
}

func main() {
    g.init()
    for {
        s:= &State{}
        s.init()
        
        // generate candidate commands (start with WAIT that never hurts)
        
        // 0. look for BUILD MINE & TOWER commands
        
        
        // 1. look at MOVE commands
        moveUnits(s)

        // 2. look at TRAIN commands
        trainUnits(s)

        // fmt.Fprintln(os.Stderr, "Debug messages...")
        fmt.Println(s.action())// Write action to stdout
    } // for
}


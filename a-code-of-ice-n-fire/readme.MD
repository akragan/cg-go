A Code of Ice and Fire
-------------------------------

Victory Conditions
* Destroy the enemy headquarters.
* After 100 turns, you have more military power than your opponent. The military power is computed by the sum of the cost of all of your units + your amount of gold

	CostTrain1 = 10
	CostTrain2 = 20
	CostTrain3 = 30

	CostKeep1 = 1
	CostKeep2 = 4
	CostKeep3 = 20

	CostMine  = 15
	CostTower = 15

	Round=0..100
	Gold 
	+1 for each active cell owned
	+4 for each active mine owned
	
	ActiveArea
	Upkeep = U3*20 + U2*4 + U1*1
	Income = ActiveArea + 4*ActiveMines - Upkeep
	
	MilitaryPower = U3*30 + U2*20 + U1*10 + Gold
	ExpectedMilitaryPower = MAX(0, U3*30 + U2*20 + U1*10 + (100-Round)*Income)
	DiscountFactor = Exp[(Round/100-1)*DiscountRate]

	DeficitToHqCapture = CheapestPathToHq-Gold
	RoundsToHqCapture = If(Income==0, 100, DeficitToHqCapture / Income)
	
	Eval = (ExpectedMilitaryPower[Me] - ExpectedMilitaryPower[Op])*DiscountFactor
	 	+ (RoundsToHqCapture[Op] - RoundsToHqCapture[Me])*(1-DiscountFactor)


Each turn:

* check chain-train-win before move

* generate BUILD MINE and TOWER commands
** build HQ mine if my income < op income
** build HQ tower when op next-chain-train-win detected
** build more towers 
*** when free area pct < 0.2
*** on op chain-train-win paths
** build 2nd mine if 
*** my income < op income
*** free area pct < 0.2

* look at MOVE commands
** iterate over all units, pick mine
** sort units to move l1, l2, l3 in this order? 
** generate candidate moves per unit with value order:
*** op hq
*** op tower / protected cell
*** op mine
*** op inactive tower
*** op inactive mine
*** op units l3/l2/l1
*** op active area
**** priority to wedge cells
**** priority to compact cells
*** op inactive area
**** priority to compact cells
*** neutral area
**** priority to compact cells
*** standing my ground
*** my cells, closer to op HQ
** sort candidate moves (for a given unit), execute

* check chain train win after move

* look at TRAIN commands



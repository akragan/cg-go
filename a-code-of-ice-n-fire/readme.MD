A Code of Ice and Fire
----------------------
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



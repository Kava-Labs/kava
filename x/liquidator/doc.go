/*
Package Liquidator settles bad debt from undercollateralized CDPs by seizing them and raising funds through auctions.

Notes
 - Missing the debt queue thing from Vow
 - seized collateral and usdx are stored in the module account, but debt (aka Sin) is stored in keeper
 - The boundary between the liquidator and the cdp modules is messy.
	- The CDP type is used in liquidator
	- cdp knows about seizing
	- seizing of a CDP is split across each module
	- recording of debt is split across modules
	- liquidator needs get access to stable and gov denoms from the cdp module

TODO
 - Is returning unsold collateral to the CDP owner rather than the CDP a problem? It could prevent the CDP from becoming safe again.
 - Add some kind of more complete test
 - Add constants for the module and route names
 - tags
 - custom error types, codespace
*/
package liquidator

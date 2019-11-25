/*
Package auction is a module for creating generic auctions and allowing users to place bids until a timeout is reached.

TODO
 - investigate when exactly auctions close and verify queue/endblocker logic is ok
 - add more test cases, add stronger validation to user inputs
 - add minimum bid increment
 - decided whether to put auction params like default timeouts into the auctions themselves
 - add docs
 - Add constants for the module and route names
 - user facing things like cli, rest, querier, tags
 - custom error types, codespace
*/
package auction

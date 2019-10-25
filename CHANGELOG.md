<!--
Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.

Usage:

Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry should ideally include a tag and
the Github issue reference in the following format:

* (<tag>) \#<issue-number> message

The issue numbers will later be link-ified during the release process so you do
not have to worry about including a link manually, but you can if you wish.

Types of changes (Stanzas):

"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"Client Breaking" for breaking CLI commands and REST routes.
"State Machine Breaking" for breaking the AppState

Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog

## [Unreleased]

## [v0.3.0]

### Features

[\#253] Add a new validator vesting account type, which releases coins on a periodic vesting schedule based on if a specific validator signed sufficient pre-commits. If the validator didn't sign enough pre-commits, the vesting coins are burned or sent to a return address.
[\#260] Pin to cosmos-sdk commit #18de630 (tendermint 0.32.6)

### Improvements

[\#257] Include scripts to run large-scale simulations remotely using aws-batch

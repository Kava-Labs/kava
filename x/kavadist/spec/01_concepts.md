<!--
order: 1
-->

# Concepts

The minting mechanism in this module is designed to allow governance to determine a set of inflationary periods and the APR rate of inflation for each period. This module mints coins each block according to the schedule such that after 1 year the APR inflation worth of coins will have been minted. Governance can alter the APR inflation using a parameter change proposal. Parameter change proposals that change the APR will take effect in the block after they pass.

Additionally this module has parameters defining an inflationary period for minting rewards to a governance-specified list of infrastructure partners. Governance can alter the inflationary period and infrastructure reward distribution using a parameter change proposal. Parameter changes that change the distribution or inflationary period take effect the block after they pass.

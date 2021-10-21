<!--
order: 4
-->

# Events

The `x/committee` module emits the following events:

## MsgSubmitProposal

| Type            | Attribute Key | Attribute Value    |
| --------------- | ------------- | ------------------ |
| proposal_submit | committee_id  | {'committee ID}'   |
| proposal_submit | proposal_id   | {'proposal ID}'    |
| message         | module        | committee          |
| message         | sender        | {'sender address}' |

## MsgVote

| Type          | Attribute Key | Attribute Value    |
| ------------- | ------------- | ------------------ |
| proposal_vote | committee_id  | {'committee ID}'   |
| proposal_vote | proposal_id   | {'proposal ID}'    |
| proposal_vote | voter         | {'voter address}'  |
| proposal_vote | vote          | {'vote type}'      |
| message       | module        | committee          |
| message       | sender        | {'sender address}' |

## BeginBlock

| Type           | Attribute Key    | Attribute Value         |
| -------------- | ---------------- | ----------------------- |
| proposal_close | committee_id     | {'committee ID}'        |
| proposal_close | proposal_id      | {'proposal ID}'         |
| proposal_close | proposal_tally   | {'proposal vote tally}' |
| proposal_close | proposal_outcome | {'proposal result}'     |

# Parameters

The kavadist module has the following parameters:

| Key        | Type           | Example       | Description                                      |
|------------|----------------|---------------|--------------------------------------------------|
| Periods    | array (Period) | [{see below}] | array of params for each inflationary period     |

Each `Period` has the following parameters

| Key        | Type               | Example                  | Description                                                    |
|------------|--------------------|--------------------------|----------------------------------------------------------------|
| Start      | time.Time          | "2020-03-01T15:20:00Z"   | the time when the period will start                            |
| End        | time.Time          | "2020-06-01T15:20:00Z"   | the time when the period will end                              |
| Inflation  | sdk.Dec            | "1.000000003022265980"   | the per-second inflation for the period                        |

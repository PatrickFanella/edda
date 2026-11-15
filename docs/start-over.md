# Start Over

Start Over removes the selected campaign’s session history, manual and automatic save points, and campaign clock in one PostgreSQL statement. If any delete fails, none of those changes commit. Other campaigns and world data remain in place. The next clock read defaults to day 1 at 08:00.

Run `go test -tags integration -run IntegrationStartOverAtomic ./internal/handlers` to check successful reset and injected PostgreSQL failures on each of the three tables. The test uses its own disposable testcontainers database and checks retained locations and another campaign’s history, saves and clock.

Reset does not restore character statistics, locations, NPCs or quests to their initial state. Coordination with an in-flight turn is separate from reset statement atomicity.

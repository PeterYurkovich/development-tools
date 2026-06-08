1. Add note to cleanup about all only being available in flag mode
2. Under resources don't place each resource in its own folder, leave everything at the top level (including perses) just move the dashboards to a folder named dashboards (note that it won't be a single file but instead 1 dashboard per file)
3. Remove the abstraction over the kubernetes client, we can just re-write later if need be, this will add complexity that isn't needed for now
4. isTui seems like it will be passed around a ton, create a context to pass around as the first parameter to functions and add it to it
5. No `obstool demo` command
6. Split the README.md###Implementation Phase into a separate TODO.md file. Remove the phases and number ordering. Update the tasks with the following format:
- Should have "tasks" and "subtask" levels
- Each should note which other item/items it is blocked by. If C is blocked by B and B is blocked by A, do not say that C is blocked by A, only say C is blocked by B.
- Be explicit about each individual command. Create atomic tasks which can be worked on in parallel
- Create a "todo" "in progress" "complete" structure (maybe [ ] [~] [x] or something similar)

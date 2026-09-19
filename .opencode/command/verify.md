---
description: Run the full verification suite (Go build/vet/test + svelte-check) and report failures
---

Run the project's full verification suite and report honestly:

1. `go build ./... && go vet ./... && go test ./...` — integration tests spin a
   PocketBase on 127.0.0.1:8035.
2. `cd ui && npx svelte-check` — the baseline is 21 pre-existing errors
   (svelte-maplibre module resolution, MapSidebar props). Diff the error list
   against that baseline and report ONLY new errors or count changes.

Report each failing step with the relevant output excerpt, and clearly
separate failures caused by the working tree from pre-existing ones. Do not
claim success while anything new fails.

$ARGUMENTS

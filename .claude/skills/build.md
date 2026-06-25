# build

Build a feature end-to-end from a brief: understand, ask what's unclear, state the plan, then implement backend + frontend in the correct repo sequence.

## Phase 1 — Parse the brief

Extract these dimensions from what the user gave you:

- **Entities**: what data is involved? Any new DB tables or columns needed?
- **Operations**: which CRUD operations (list, get, create, update, delete)?
- **Access control**: auth required? Company-scoped? Admin-only? Public?
- **UI**: new page? Component on existing page? Which route under `/app/`?

## Phase 2 — Ask before writing a single line of code

Identify every ambiguity or missing decision. If ANY of these is unclear, ask the user now. Batch all questions into one message using `AskUserQuestion` — do not ask one at a time.

Common things to ask:
- Does this need a DB migration, or does it reuse existing tables?
- New feature module (`internal/<feature>/`) or extension of an existing one?
- Company-scoped (`RequireSelectedCompany`) or user-scoped (`RequireAuth`)?
- Pagination / filtering on list endpoints?
- New sidebar nav entry, or linked from an existing page?

Do not guess at scope. Asking now is cheaper than reimplementing.

## Phase 3 — State what you'll build

Before touching any file, write a short plain-text summary of exactly what you'll create:
- Which backend module, which endpoints, any new migration
- Which frontend route, whether a sidebar entry is needed

This gives the user one last chance to redirect before implementation begins.

## Phase 4 — Implement (in this exact order)

### Backend

**1. Migration** — if new table or column is needed:
```bash
make migrate name=<snake_case_description>
```
Then write the SQL in both generated up/down files.

**2. sqlc queries** — add SQL queries to the feature's query file; run `make gen` after to regenerate `internal/platform/compiled/` and the TypeScript SDK.

**3. Types** (`internal/<feature>/types.go`):
```go
type CreateFooRequest struct {
    Name string `json:"name"`          // camelCase JSON tags — enforced by make check
}
type FooInfo struct {
    ID        int64  `json:"id"`
    Name      string `json:"name"`
    CreatedAt string `json:"createdAt"`
}
type CreateFooInput struct{ Body CreateFooRequest }
type CreateFooOutput struct{ Body FooInfo }
```

**4. Handler** (`internal/<feature>/handler.go`):
```go
func (h *Handler) handleCreate(ctx context.Context, input *CreateFooInput) (*CreateFooOutput, error) {
    p, err := humax.RequireAuth(ctx)  // or RequireSelectedCompany for company-scoped
    if err != nil {
        return nil, err
    }
    // ... h.queries.FooCreate(ctx, ...)
    return &CreateFooOutput{Body: toFooInfo(row)}, nil
}
```

**5. Setup** (`internal/<feature>/setup.go`):
```go
huma.Register(deps.API, huma.Operation{
    OperationID: "create-foo",   // kebab-case, unique across all routes
    Method:      http.MethodPost,
    Path:        "/api/foos",
    Summary:     "Create a foo",
    Tags:        []string{"foo"},
    Security:    humax.BearerAuth(),
}, h.handleCreate)
```

For a **public endpoint** (no auth): omit `Security: humax.BearerAuth()` and don't call `RequireAuth` in the handler.

**Path parameters**:
```go
type GetFooInput struct {
    ID int64 `path:"id"`
}
// Register with Path: "/api/foos/{id}"
```

**Query parameters**:
```go
type ListFoosInput struct {
    Limit  int `query:"limit"`
    Offset int `query:"offset"`
}
```

**6. Wire the module** — only if this is a new module (not extending an existing one):
- `cmd/server/main.go`: add `foo.Setup(ctx, deps)`
- `internal/platform/routes/routes.go`: add `foo.Setup(ctx, deps)`

**7. Regen** — if not already done in step 2:
```bash
make gen   # updates openapi.json + frontend/src/lib/api/
```

### Frontend

**8. Page** — `frontend/src/routes/app/<name>/+page.svelte`:
```svelte
<script lang="ts">
    import { createQuery, createMutation, useQueryClient } from '@tanstack/svelte-query';
    import { toast } from 'svelte-sonner';
    import PageHeader from '$lib/components/layout/page-header.svelte';
    import { listFoos, createFoo } from '$lib/api-client.js';   // generated SDK
    import type { FooInfo } from '$lib/api/types.gen.js';

    const qc = useQueryClient();
    const query = createQuery({
        queryKey: ['foos'],
        queryFn: () => listFoos().then((r) => r.data?.foos ?? [])
    });
    const mut = createMutation({
        mutationFn: (body: { name: string }) => createFoo({ body }),
        onSuccess: () => { qc.invalidateQueries({ queryKey: ['foos'] }); toast.success('Created'); },
        onError: (e: Error) => toast.error(e.message ?? 'Failed')
    });
</script>

<PageHeader title="Foos" description="Manage your foos" />
<main class="px-6 py-4">
    {#if $query.isLoading}<!-- skeleton -->{/if}
</main>
```

**9. Sidebar** — add to `frontend/src/lib/components/layout/app-sidebar.svelte` if it's a new top-level page:
```ts
{ href: '/app/foos', label: 'Foos', icon: SomeIcon }
```

**PageHeader with an action button**:
```svelte
{#snippet actions()}
    <Button onclick={() => (open = true)} size="sm">
        <Plus class="mr-1 h-4 w-4" /> New Foo
    </Button>
{/snippet}
<PageHeader title="Foos" description="…" actions={actions} />
```

**Auth context in frontend**:
```ts
import { getProfile } from '$lib/auth.js';
const profile = getProfile();
// profile?.userId, profile?.selectedCompanyId, profile?.name
```

### Available UI components

Import from `$lib/components/ui/`:

| Component | Import path |
|-----------|-------------|
| Button | `$lib/components/ui/button/index.js` |
| Input, Label | `$lib/components/ui/input/index.js`, `$lib/components/ui/label/index.js` |
| Badge | `$lib/components/ui/badge/index.js` |
| Table, TableBody, TableCell, TableHead, TableHeader, TableRow | `$lib/components/ui/table/index.js` |
| Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter | `$lib/components/ui/dialog/index.js` |
| Card, CardHeader, CardTitle, CardContent | `$lib/components/ui/card/index.js` |
| Skeleton | `$lib/components/ui/skeleton/index.js` |
| Separator | `$lib/components/ui/separator/index.js` |

### Verify

**10.** `make check` — go vet + JSON tag lint + tsc (target: 0 errors)
**11.** `cd frontend && bun run check` — Svelte type check (target: 0 errors)

## Hard rules

| Rule | Why |
|------|-----|
| Never edit `frontend/src/lib/api/` | Regenerated by `make gen` — edits will be overwritten |
| JSON tags must be camelCase | `make check` rejects snake_case tags |
| Feature packages must not import each other | `make check` enforces no cross-domain imports |
| Auth per-handler only | No global auth middleware; each handler calls `RequireAuth` or `RequireSelectedCompany` |
| Svelte 5 syntax | `$state()`, `$derived()`, `$props()` — not `let`/`export let` reactive declarations |
| No self-closing `<div />` | Svelte requires `<div></div>` |
| `$derived` in query options | Wrap derived values in a getter: `get enabled() { return $derived(...) }` |
| SDK imports | From `$lib/api-client.js`, not directly from `$lib/api/` |

## Error helpers

```go
humax.NotFound("not found")       // 404
humax.Conflict("already exists")  // 409
humax.Forbidden("not allowed")    // 403
humax.Unprocessable("bad state")  // 422
humax.BadRequest("bad input")     // 400
humax.Unauthorized("not logged in") // 401
```

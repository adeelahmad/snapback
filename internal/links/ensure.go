package links

import "context"

// Ensure creates the link for dir if absent, or proves an existing one is owned.
func (e *Engine) Ensure(ctx context.Context, dir string) (Result, error) {
	panic("SUB-AGENT-TODO: T4 — SPEC §8 steps 1-6: honour ctx; place; openDirChain; caseFoldSibling check; lstatAt; absent -> Put pending, symlinkAt, Put owned, Created=true; present and proven owned (record + exact target) -> Created=false; otherwise link_conflict, entry untouched; on EEXIST re-lstat and apply the same proof; EACCES/EROFS -> permission_denied with pending record rolled back; never touch network or provider")
}

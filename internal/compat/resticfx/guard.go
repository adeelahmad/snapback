package resticfx

// Guard holds the injected roots GuardRepo checks against.
type Guard struct {
	TempRoot string
	Home     string
}

// GuardRepo refuses any repository that is not a disposable temp path or the Stage 1 rclone remote.
func GuardRepo(repo string, g Guard) error {
	panic("SUB-AGENT-TODO: allow absolute path strictly under TempRoot (Clean + EvalSymlinks where exists) or exactly rclone:gdrive:snapback-stage1; refuse relative, /, TempRoot itself, Home outside TempRoot, other rclone:/backend specs")
}

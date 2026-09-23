package crash

// RecoverAndReport installs the panic-only crash-reporting recover hook.
// Call it directly inside a defer statement so recover() runs in the
// deferred function it returns, the only place recover() is effective:
//
//	defer crash.RecoverAndReport(opts)()
//
// On a real panic, the returned function recovers it, builds an Envelope
// from the recovered value and this goroutine's stack, calls Report with
// opts, then re-panics with the exact original value so the process still
// crashes exactly as it would without this hook -- same exit code, same
// stderr, just with a Report side-effect first. A normal return, a normal
// error value, or a clean exit are reported nowhere. There is no way to
// intercept log.Fatal: it calls os.Exit directly and bypasses every defer
// in the process, so this hook's recover() is simply never invoked on that
// path -- no special-casing is possible or needed.
func RecoverAndReport(opts Options) func() {
	panic("SUB-AGENT-TODO: implement RecoverAndReport (S6-08/T4)")
}

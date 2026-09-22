# demo.tape — the quick start GIF for the docs site.
#
# The commands below are exactly the counted action path of
# test/acceptance/testdata/actions-linux.sh (install, setup, ls .snapshot) plus
# the restore it earns. test/site/demo_tape_test.go pins them to that script, so
# the GIF cannot drift from the path the acceptance test counts.
#
# Record with: vhs docs-site/img/demo.tape

Output docs-site/img/demo.gif

Set Width 1200
Set Height 600
Set FontSize 20
Set Padding 24
Set Theme { "background": "#202124", "foreground": "#f8f9fa", "cursor": "#f8f9fa", "selection": "#3c4043" }
Set TypingSpeed 60ms

Type "curl -fsSL https://snapback.run/install.sh | sh"
Enter
Sleep 6s

Type "cd ~/project && snapback setup"
Enter
Sleep 8s

Type "ls .snapshot"
Enter
Sleep 4s

Type "cp .snapshot/latest/report.docx ."
Enter
Sleep 4s

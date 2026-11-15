# Screen-fitting composer (#62)

The narrative panel remains mounted while details are open. Draft text, textarea
selection and transcript DOM state survive a return to play. Campaign changes
mount a separate workspace, so drafts cannot leak between campaigns.

Choices, multiline input and error/pending feedback each scroll within a bounded
region. The send row remains outside those scroll regions. Existing choice
buttons still submit directly, and free text keeps Enter/Shift+Enter behavior.
Only an accepted dispatch clears the draft. Empty drafts cannot be submitted.

Verification covers the six default roadmap viewports, eight long choices,
long history, a large multiline draft, draft selection, expanded saves, enlarged
text and deterministic visual-viewport resizing. Automated resizing is not a
real software-keyboard check; #46/#78 retain that device qualification gate.
The existing palette, font families, square borders and focus treatments remain.

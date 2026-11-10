# The Aftersight pipeline in GoRex

GoRex is the project `gorex` on the Aftersight server. Its seeds are `.aftersight/project.json`
(tests, and the paths each role may write) and `.aftersight/claude.json` (the Claude pipeline's
release, stages and outside-review gate). A change to either reaches runs only after
`aftersight config push` from a commit that holds it.

## The outside-review gate needs `OPENAI_BASE_URL`

At `second-opinion` rigor and above, `dual-review.sh` sends the brief and the diff to an outside
model. `.aftersight/claude.json` names that model (`glm-5.3`), its provider (`zai`) and the
variable that holds the key (`ZAI_API_KEY`). The endpoint is not a gate setting in this release:
the server refuses `gate.OPENAI_BASE_URL`. The gate reads it from the environment of the Claude
session that runs the pipeline. So every machine that runs the pipeline sets:

    OPENAI_BASE_URL=https://api.z.ai/api/coding/paas/v4

for example in the `env` block of `.claude/settings.local.json`, which git ignores. The gate has no
default endpoint: without the variable it stops before sending anything, and pre-flight fails the
rows `dual-review.sh interface` and `gate model (outside-review) credentialed + reachable`.
`ZAI_API_KEY` must be set in the same environment.

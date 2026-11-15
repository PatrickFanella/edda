# Supported Fighter advancement

Issue #73 adds one transition: the supported SRD 5.2.1 Fighter from level 1 to
level 2 at 300 earned XP. The sheet offers fixed HP (6 + Constitution, minimum
1) or a d10 roll + Constitution (minimum 1). Preview is inert; confirmation
saves the choice and result, level, maximum HP, additional d10 Hit Die and
history together. A replay after reload reads the saved result without another
roll or increase. If the transaction fails, no result is committed or shown; a
retry can roll again. Current HP and spent Second Wind uses do not reset.

The transition retains Defense and the three recorded weapon mastery choices.
Replacing them, multiclassing and later levels are unavailable. Proficiency
stays +2; the next XP threshold is 900. Action Surge (one use) and Tactical Mind
are recorded and shown as unavailable effects, as the existing level-one
feature effects are. This is a progression fixture, not full Fighter support.

XP must already have been earned and recorded by a trusted source. This PR does
not add an XP award button or infer rewards from narrative or defeated targets;
encounter XP remains #149. Legacy XP, skill points, abilities and tools are not
converted; the legacy mutation guard still applies to versioned campaigns.
There is no migration or production data mutation.

Sources: the [pinned SRD 5.2.1](srd-5.2.1-conformance.md), level advancement and
fixed HP p. 23, Fighter table and retain/replace fighting style p. 47, mastery,
Action Surge and Tactical Mind p. 48. This is an adaptation under the attribution
and license linked there.

Rules tests exercise fixed/rolled HP, threshold and choice rejection and
unsupported further advancement. Isolated PostgreSQL tests exercise rollback
on history failure, idempotent receipt replay, reload, stale requests and
unchanged unrelated stats/skill points. Browser fixtures exercise desktop and
phone choice, preview, confirmation and refreshed level/HP. These tests do not
qualify XP award generation, live providers, real device keyboards or deployment.

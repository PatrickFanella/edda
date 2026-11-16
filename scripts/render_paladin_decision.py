#!/usr/bin/env python3
"""Render the reviewed source/native receipt as a readable decision, without network writes."""
import json
from pathlib import Path

ROOT=Path(__file__).resolve().parents[1]
DIRECTORY=ROOT/'docs/investigations'


def main():
    m=json.loads((DIRECTORY/'issue-327-paladin-scope.json').read_text())
    r=json.loads((DIRECTORY/'issue-327-link-verification.json').read_text())
    by={x['key']:x for x in m['published_issues']}
    sealed=m.get('native_validation',{}).get('supplements')=='sealed'
    text=f'''# Paladin and Oath of Devotion decision (#327)

Paladin and Devotion remain unqualified. This decision publishes 121 Task issues: 54 class execution/admission/application leaves, 54 finite shared spell leaves, five condition leaves, two Dispel Magic leaves and six qualification gates. It changes no runtime, creation eligibility, grants, derived bonuses, coverage or deployment. Closing #327 does not implement any of these leaves.

The read-only native receipt was observed at {r['observed_at']}: {len(r['graph'])} issues, with exact title/body/parent/prerequisite and Project Task/Size readback for all 121 new issues. The full graph and metadata are in [the receipt](issue-327-link-verification.json); all 87 source records, 186 distinct mechanic/option keys, nested choices and numeric operands are in [the matrix](issue-327-paladin-scope.json). The matrix distinguishes 195 direct required owners from the complete 352-node transitive ownership path, including source metadata and planning prerequisites. Those counts are ownership records, not claims of implemented effects.

Parent supplements are **{'SEALED' if sealed else 'PENDING root serial publication'}**. The exact marked sections are in [the supplement artifact](issue-327-parent-supplements.json). The #47 index is 60 characters; the other sections are below 750. The immutable pre-Paladin prefixes come from root-accepted Monk merge `3042172c935bdefa606c008c8ddda277b7548fd8`. Their complete bodies and independently pinned hashes are retained. Publication must preserve each complete accepted prefix plus the exact owned section; rollback, changed or duplicate markers, and altered second owned sections fail. Later unrelated pure appends pass. No shared parent body was edited by this task.

## Source and implementation evidence

The pinned English SRD 5.2.1 PDF SHA256 is `8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`. Paladin is PDF 53–55, spell membership 55–56 and Devotion 56–57. Consumed rules include advancement 22–25, Cleric cantrips 38, Holy Symbol 97, casting 104–106, Cover 179, Blinded 177, Charmed 178, Deafened 181, Frightened 182, Incapacitated/Invisible 184, Long Rest 185, Paralyzed/Poisoned 186, Short Rest 187 and Stunned 189. Each spell leaf names its exact spell pages and source operands. CC-BY-4.0 attribution and adaptation: [source policy](../srd-5.2.1-conformance.md). Fixtures and ownership boundaries are Edda additions.

At base `c577e3e7a8dac94d579b65f58af444fe427a412e`, `NewSupportedBuild` and `Project` admit the fixed Fighter/Cleric fixtures; `prepareAdvancement` supports Fighter 1→2 at 300 XP. Paladin form labels and the narrative Oathbound Emblem do not establish Paladin rules admission or starting equipment. `ValidateReadiness` rejects unimplemented catalogue requirements. Existing engine confirmation, game previews, HTTP sheet routes, web confirmation UI and trusted TUI commands provide integration seams; they supply no Paladin feature execution. The matrix records exact inspected file hashes and findings, including handlers and both clients. No source catalogue or checklist is presented as executable support.

## Native ownership and qualification

Paladin top gate #845 is a native child of #327, preserving #283 capacity. It requires decision #327, class gate #846 and shared gate #847, and directly blocks #337. Class gate #846 requires all 54 class leaves. Shared gate #847 requires the finite cohorts and exact existing/shared owners. Cohorts #848/#849/#850 are native children of existing #277 under #191; they contain exactly 20/20/14 spell leaves. Each cohort directly blocks #279 and #280. Admissions #960/#961/#962 are under #268; trusted endings #963/#964 and Dispel follow-ups #965/#966 are under #193. All seven also directly block #279/#280, as their source contracts require. Every new implementation leaf is mandatory through these paths even after #327 closes. Containment and blocking DAGs are checked separately.

The sole new direct #429 consumer is Paladin's own D10/fixed6 adapter #851. All 20 applications and later-class entry require it, so their mandatory path reaches #429→#336/#285 without redundant gate/row edges. The 98 earlier #429 consumers and all 229 recorded #285/#336/#429 edges remain intact. Ranger's unrelated #968 gateway subsequently filled #429's 100th outgoing slot; it is not a Paladin prerequisite. The verifier permits 100 and rejects 101, and also checks incoming and child limits. #191 gained no direct children. Native shared qualification incoming counts after this publication are #279=97 and #280=92; these are observations, and unrelated future prerequisites may be added while every recorded edge and exact shared body remains protected.

Feature execution, first-class creation, later-class eligibility and each level application are separate contracts. Paladin level selects features, prepared counts, slots, resource capacities and its HP adapter. Canonical **total character level** independently selects XP and live PB. The unequal witness is Fighter 4/Paladin 1→Paladin 2: total 6 requires 14,000 XP and PB+3, while Paladin 2 still supplies its own features/slots/HP. It must not use advancement row 2's 300 XP/PB+2. Later entry Fighter 4→Fighter 4/Paladin 1 uses total 5, 6,500 XP/PB+3 and later-level HP, not first-character 10+CON. Boolean/fractional levels, total 21 and unqualified current classes fail closed. Every named feature's source keys and granting-level path are checked independently; ASI owns no unrelated feature keys.

## Every Paladin row

These are source requirements, not successful runtime fixtures. The PB column is single-class table metadata, never a multiclass eligibility operand. Prepared counts exclude always-prepared spells. Slots are levels 1/2/3/4/5. Devotion grants are conditional on the separately confirmed subclass.

|Level|Native application|Table PB|Channel uses|Prepared|Slots 1/2/3/4/5|Lay On Hands|Named source grants|
|---|---|---|---|---|---|---|---|
'''
    prepared=next(x['record'] for x in m['records'] if x['record']['id'].endswith('choice/paladin/prepared-spells'))
    for row in m['levels']:
        n=row['level'];v=row['values'];p=next(x['value'] for x in prepared['values_by_level'] if x['level']==n)
        grants=list(row['grants'])
        grants+={3:['Devotion spells and Sacred Weapon if selected'],5:['Devotion spell row if selected'],7:['Aura of Devotion if selected'],9:['Devotion spell row if selected'],13:['Devotion spell row if selected'],15:['Smite of Protection if selected'],17:['Devotion spell row if selected'],20:['Holy Nimbus if selected']}.get(n,[])
        text+='|'+'|'.join(map(str,[n,'#'+str(by[f'level-{n}']['issue']),row['progression']['proficiency_bonus'],v.get('channel-divinity',0),p,'/'.join(str(v.get(f'spell-slot-{j}',0)) for j in range(1,6)),v['lay-on-hands'],', '.join(grants) or 'No new named feature; apply exact count changes']))+'|\n'
    text+='''
## Shared effects and source boundaries

#503 remains four-class membership only; #856 owns the exact 38-spell Paladin list, independently of execution. #855 owns Paladin Charisma/focus routing, #857 preparation and #858 five slot columns. Blessed Warrior uses only the seven Cleric cantrips from #503's domain and its own Paladin-Charisma adapter. #58 remains sheet projection and #273 remains casting modes. Holy Symbol never substitutes costly/consumed materials. Prepared spell levels remain class-specific despite combined slots. Recommendations are not automatic choices.

Existing #352 owns bounded level-one Bless; #905 extends only its missing upcast/eligibility integration. #355 owns low-level creature Dispel Magic; #965/#966 implement the two exact remaining #359 scopes, rather than treating that spike's closure as execution. Detect Magic's #265 free-use contract remains unchanged; its class-slot/ritual extension is separately published. Cure Wounds consumes #263/#615 with the Paladin CHA adapter, without importing #263's fixed WIS caster eligibility.

The source-correct Light extension #967 is required by Blessed Warrior and the spell cohort. Historical #264 alone is insufficient: its willing-other-holder targeting conflicts with PDF 144. The new owner preserves valid self-carried/duration/storage behavior, delegates the common effect store and #145 adjudication, and enforces one Large-or-smaller eligible object, Touch/V+M, chosen color, one hour, opaque-cover blocking, 20/20-foot light and recast ending. Objects worn/carried by someone else remain invalid even if that holder is willing. PDF 106 separately permits the caster to dismiss its own ongoing time-span spell without an action if not Incapacitated; #967 consumes #487 for that predicate. Incapacitated alone does not end Light. #264's historical body/prerequisites stay unchanged. Detect Magic #906 explicitly consumes the PDF 104 prepared Ritual rule, normal casting time plus ten minutes without a slot or extra Ritual Adept requirement, and PDF 105's longer-casting action and Concentration rules.

#488 supplies persisted Charmed/Frightened ending/immunity APIs, not merely projection. #504 retains only Paralyzed/Poisoned/Stunned ending, and #624 is unchanged. Restoring Touch owns only its Paladin source/target/cost request and consumes actual endings, including #963/#964 for Blinded/Deafened. #625 supplies the existing curse/Attunement seam; Remove Curse preserves its ALL-versus-ONE selection distinction and leaves a cursed item's curse intact. New spell-owned possession, planar and summon leaves must implement their complete bounded authoritative transitions. Generic #185/#274 planning labels never substitute for those implementations. Unknown world facts, source variants and effects remain unavailable.

Original #334/#390 guards, all recorded historical #337 edges, prior class owners and shared #779's exact body/prerequisites remain protected. #779's body SHA256 remains `556d1618d1cd2bd9c956fe38861d89e94ee14ff61a505b5d2b6a1fa46d19d150`. Every required terminal specifies source keys, numeric eligibility, authoritative inert preview/explicit confirmation, targets/range/timing/cost/recovery/expiry/alternatives, independently violated negatives, durable replay/reload/concurrency and atomic rollback, web and trusted TUI confirmation, and isolated PostgreSQL verification. Passive/source-automatic benefits use the triggering event; no optional activation is invented.

## Exact published inventory

|Issue|Native parent|Task Size|Contract|
|---|---|---|---|
'''
    for x in sorted(m['published_issues'],key=lambda x:x['issue']):
        text+=f"|#{x['issue']}|#{x['parent']}|{x['project_size']}|{x['title']}|\n"
    text+='\n## Class execution and application decisions\n\n'
    for x in m['published_issues']:
        if x.get('category')=='class':text+=f"- **#{x['issue']}: {x['title']}.** {x['fixture']}\n\n"
    text+='''## Verification and limits

Passed once: `GOMAXPROCS=2 go test -p 2 -short -race ./...` with unchanged default timeouts/assertions (SRD package 238.278s), and `GOMAXPROCS=2 go vet -p 2 ./...`. No further Go/vet full run was performed. An initial invalid fused `-p2` invocation stopped at package selection before tests ran; the valid command above was the sole full suite. The known Go1.27.3 golangci-lint shim has no binary; no installation/toolchain/host changes were made.

Pinned Paladin, Devotion, advancement, multiclass and all four feat comparators pass. The additional consumed-spell comparator checks all 50 spell envelopes and 84 numeric/alternative source clauses, rejecting 50 contrary envelopes, 84 contrary operands and five inherited general-casting contradictions. Independent checks compare the actual #906/#967 candidate contracts with those source rules. It reads the pinned PDF directly, independently of the decision's source transcription. Ten focused Paladin Python tests pass with the pinned PDF, including inherited ritual/dismissal controls, class applications and immutable parent markers; the 16 existing Fighter guard tests and two pagination tests also pass. Native probes independently remove every new owned prerequisite and each recorded protected shared edge, corrupt bodies/parents/titles/Task Size, test separate cycles and all three 100 limits, and permit unrelated future #279/#280 prerequisites. Sealed probes additionally reject rollback and altered/duplicate owned sections while accepting unrelated pure appends.

Run `python3 scripts/verify_paladin_scope.py --native --pdf /tmp/edda-wave1/SRD_CC_v5.2.1.pdf --probes`; add `--live` for one bounded read-only receipt refresh. Final parent publication acceptance adds `--sealed --require-publication`. REST/GraphQL pagination has finite call caps and repeated-page/cursor rejection; exact metadata queries reject truncated Project results. No live runtime/database, integration-tag PostgreSQL execution, browser/device journey, production admission, inference or deployment is claimed. Those implementation proofs remain obligations of the published leaves. Parent owns serial supplements, fresh independent acceptance, final CI and merge. The PR stays draft and closes #327 only; no watch, polling, merge or issue closure was performed here.
'''
    (DIRECTORY/'issue-327-paladin-decision.md').write_text(text)
    print('Rendered20row decision and121issue native inventory; supplements '+('SEALED' if sealed else 'PENDING'))


if __name__=='__main__':main()

#!/usr/bin/env python3
"""Render the actual immutable Rogue publication roster without rebuilding its source audit."""
import json
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]
OUT=ROOT/'docs/investigations'

def render(m):
    by={x['key']:x for x in m['published_issues']}
    lines=['# Rogue and Thief decision (#329)','',m['publication']+'.','',
      'This decision supplies a finite implementation roadmap and native qualification paths. It implements no Rogue runtime. Universal source identity/eligibility and typed-dispatch permission are distinct from qualified effect execution. Every unknown/unqualified output remains unavailable. Full production class qualification consumes both #279 and #280 as well as the Rogue feature/level cohorts; broad Rogue and Full Rules remain unavailable.','',
      'Pinned English SRD5.2.1: Rogue PDF61–63,Thief64,advancement23,multiclass24–26; consumed source locators appear in each terminal. SHA256 `'+m['source_sha256']+'`. Complete canonical records retain CC-BY-4.0 attribution/adaptation in their source headers and `docs/srd-5.2.1-conformance.md`. The source contracts companion contains direct complete consumed spell/item contracts,not selected quotations. Ownership/fixtures are Edda adaptations.','',
      '## Native qualification architecture','',
      'Qualification #1093 is under #329 and remains blocked by #329,#279,#280,feature cohort #1094,level cohort #1095 and shared cohort #1096. Cohorts #1094/#1095 sit under #1093; shared #1096 sits under #192. Hide #1137 sits under #50; other shared items sit under #1096. Execution dependency gateway #1097 sits under #329,consumes only #61/#64/#65/#66/#67/#273/#353,and has an exact finite consumer roster. It supplies no runtime prerequisites or implementations. Own HP adapter consumes #968→#429; no new direct #429 edge.','',
      'No child is added under #283. One shared cohort edge is added to each of #279/#280,and one qualification edge to #337. Dependency and containment DAGs and all three100-link caps are checked separately. Every preexisting required path is retained; all owned paths remain mandatory after #329 closes. Four parent supplements are local marked artifacts for ROOT publication only. Accepted POST-RANGER prefixes were freshly matched and frozen; ROOT published all four exact sections and the fresh FINAL sealed proof passed.','',
      '## All twenty class rows','',
      'Class PB is source metadata,not the multiclass XP/PB operand. Canonical live advancement selects the complete row by authoritative post-transition TOTAL admitted class levels. Fighter4/Rogue1→Rogue2 reaches total6,14000XP/PB3; Fighter4→Rogue1 reaches total5,6500XP/PB3. Rogue feature eligibility,dice and resources remain Rogue-class-level. Reject totals>20,invalid/noninteger levels,nonadjacent applications and unqualified class vectors.','',
      '| Rogue level | Source PB | Sneak Attack | Applied feature owners | Application |','|---|---|---|---|---|']
    for n in range(1,21):
        row=next(x['record'] for x in m['records'] if x['record']['kind']=='level' and x['record']['level']==n)
        lines.append(f"|{n}|{row['progression']['proficiency_bonus']}|{(n+1)//2}d6|"+', '.join(m['grants'][str(n)])+f"|#{by[f'level-{n}']['issue']}|")
    lines+=['','Later-class first entry has its own application #'+str(by['later-1']['issue'])+' and actual own eligibility adapter. First-character branches consume creation traits/skills/equipment,without adding multiclass eligibility as a runtime prerequisite.','','## Actual runtime evidence','']
    lines += [f"- `{p}`: {e['finding']}" for p,e in m['evidence'].items()]
    lines+=['','## Exact one-PR publication roster','','Every Task is Size M. The full body/source-key/prerequisite/parent/metadata receipt is `issue-329-link-verification.json`; the durable mutation journal is `issue-329-publication-journal.json`. Gate closure is qualification,not behavior implementation.','']
    for x in m['published_issues']:
        lines += [f"### #{x['issue']}: {x['title']}",'',x['fixture'],'',f"Native parent #{x['parent']}; exact blockers: "+', '.join('#'+str(n) for n in x['blocked_by_issues'])+'.','']
    lines+=['## Shared confirmation and client contract','',m['common_acceptance'],'',
      'All leaves require positive and independently violated numeric/source/target/timing/eligibility controls,isolated PostgreSQL history/effect/choice rollback,durable same-ID replay/reload/concurrency and explicit shared web AND TUI availability. Short tests are separate from integration-tag PostgreSQL,desktop/phone/browser/device qualification. No production access,inference,deployment or runtime promotion follows from publishing this roadmap.','',
      '## Complete source-record execution matrix','','Each canonical record,including every nested option/prerequisite/resource,is retained in `issue-329-rogue-scope.json`. Its `key_owners` maps EVERY required mechanic and option key; each row also names hashed engine/web/TUI evidence. All are currently metadata-only/unqualified.','',
      '| Complete source record | Exact execution owners |','|---|---|']
    for x in m['records']:
        lines.append('| `'+x['record']['id']+'` | '+', '.join(f"{k} (#{by[k]['issue']})" for k in x['execution_owners'])+' |')
    lines+=['','## Universal identity permission and conditional effects','',
      'The complete597-record inventory (339spells,258magic-item headings) is retained in `dispatch_inventory`,including source pages/IDs and every unqualified dispatch ID. Universal Spell Scroll identity permission does not grant Spellcasting,prepared spells,slots or effect execution. All0–9 scroll source levels have exact DC/attack-table contracts; Thief’s reliable0/1 and INT(Arcana) checks for EACH2–9 are separately tested. Unknown/unqualified effects are rejected before rolling/spending. Qualified failed higher-level Thief casts disintegrate the scroll with no effect. Normal interrupted item casting preserves it. Ordinary scroll failure removes its spell without invented physical destruction.','',
      'Concrete spell composition fixtures consume Light #967,Detect Magic #906 and Aid #614; Ring of Jumping consumes Jump #1064. These are actual existing owner contracts,not copies of their effects. Four distinct attunable item effects and Wand dawn recharge are independent leaves. Spell Recall #399 remains unavailable to pure Rogue,Magic Initiate or UMD-only actors; actual authoritative qualified Spellcasting is required. Ordinary feat selection adds no blanket caster,encounter or Concentration condition.','',
      f"Remaining {len(m['unqualified_dispatch_ids'])} source identities lack a qualified effect binding through this Rogue dispatcher; their COMPLETE names/IDs/source pages remain explicit in the JSON. This is a dispatch availability statement,not a claim that other shared decisions lack owners. Their source presence does not qualify them.", '',
      '## Proof boundaries','',
      'The strict offline/live verifier checks complete source transcriptions,source-key ownership,full actual native contracts,Task/Project Size,all old required paths,all own paths after decision closure,exact finite gateway rosters,both DAGs and native caps. Independent semantic reconstruction and immutable seals reject coherently weakened own matrices/receipts. FINAL parent proof separately rejects missing release,original-only rollback,changed prefixes/sections and duplicate or altered duplicate markers,while accepting later unrelated parent appends. No current live body is automatically accepted as a released prefix.']
    if m.get('final_seal')=='PASS':
        lines+=['','## Dated publication and final seal','','The four `-PENDING.md` files are immutable prepared artifacts. Their embedded PENDING wording records the earlier handoff stage; ROOT published their exact bytes on2026-11-16 at20:06:53–20:07:04UTC. The final sealed observation at '+m['final_proof']['observed_at']+' passed and supersedes that workflow status. The artifacts stay byte-for-byte unchanged for independent prefix/section/marker verification. This seals publication evidence,not runtime implementation.']
    # Use actual Hide ID rather than a guessed sequence number in prose.
    text='\n'.join(lines)+'\n';text=text.replace('Hide #1137',f"Hide #{by['hide']['issue']}")
    (OUT/'issue-329-rogue-decision.md').write_text(text)

if __name__=='__main__':render(json.loads((OUT/'issue-329-rogue-scope.json').read_text()))

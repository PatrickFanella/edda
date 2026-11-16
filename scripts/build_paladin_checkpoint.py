#!/usr/bin/env python3
"""Build the unpublished Paladin audit. No network calls or tracker writes."""
import hashlib
import json
from pathlib import Path
from paladin_shared_plan import shared_tasks

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / 'docs/investigations'
PIN = '8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87'


def build():
    destination=OUT/'issue-327-paladin-scope.json'
    previous=json.loads(destination.read_text()) if destination.exists() else {}
    catalogues = {f'internal/rules/srd521/testdata/character-{name}.json': json.loads(
        (ROOT / f'internal/rules/srd521/testdata/character-{name}.json').read_text())
        for name in ['paladin', 'devotion', 'advancement', 'multiclass']}
    records = [r for p, c in catalogues.items() if p.endswith(('character-paladin.json', 'character-devotion.json')) for r in c['records']]
    tasks = []

    def task(key, title, fixture, prerequisites=(), depends=(), size='M'):
        tasks.append(dict(key=key, title=title, fixture=fixture, prerequisites=list(prerequisites),
                          depends=list(depends), project_size=size, issue=None, publication='UNPUBLISHED'))

    task('hp', 'Apply Paladin D10 or fixed-six HP through shared advancement',
         'PDF22–25/53. Paladin-specific eligibility. First-character HP10+CON; subsequent Paladin level adds max(1,d10+CON) or max(1,6+CON), one D10 Hit Die. Total-level XP/PB and retroactive CON correction consume429, never Fighter731 eligibility. Positive: CON+2 fixed increase8; rolled1 with CON-3 increases1; CON17→18 at total8 adds8 retroactively in addition to new-level HP. Negative: D12/fixed7/D8/fixed5, first-level10 on later-class entry, rerolled saved die, missing XP. Exactly ONE new direct429consumer; all other paths use this adapter.', [429, 58])
    task('traits', 'Confirm first-class Paladin traits and two skills',
         'PDF53/22. STR and CHA primary; WIS/CHA saves; Simple/Martial weapons; Light/Medium/Heavy armor and Shields. Exactly2 distinct Athletics/Insight/Intimidation/Medicine/Persuasion/Religion skills. Positive: Athletics+Religion, independent source provenance. Negative: three skills, duplicate skill, Survival, unqualified species/background, later-class receiving these saves or skills.', [430, 333, 58])
    task('equipment', 'Confirm either Paladin starting equipment package',
         'PDF53/92–99. A: Chain Mail, Shield, Longsword,6 Javelins,Holy Symbol,Priest’s Pack,9GP; B:150GP. Positive: either bundle exactly once. Negative: mixed bundles,8 Javelins,155GP, missing Holy Symbol, spending uncommitted gold. Consume qualified item/equip contracts; this leaf grants inventory only and copies no item effects.', [267, 430])
    task('multiclass', 'Validate Paladin later-class Strength and Charisma eligibility',
         'PDF24–25/53. STR13 AND CHA13 plus every current-class primary13. Gain D10,Martial weapons,Light/Medium armor/Shields and level1 features only; no Heavy armor,Simple proficiency,save/skill/equipment regrant. Positive: STR13/CHA13 and current-class threshold13; independently reject12 in each. Combined caster slots add ceil(Paladin level/2); prepared spells independently determined as single-class Paladin. Extra Attack does not stack.', [335, 430], ['hp'])
    task('casting', 'Route Paladin Charisma and Holy Symbol casts through receipts',
         'PDF54/97/104–106. Attack=PB+CHA; DC8+PB+CHA. Holy Symbol substitutes only noncostly/nonconsumed M, never costly/consumed M; V/S/hands/armor,one slot spell per turn,cast timing,range and target eligibility revalidated. Positive: CHA+3/PB2 gives DC13/attack5. Negative: WIS adapter,untrained armor,occupied required hand,unprepared spell,unknown effect. #273 modes and #58 projection do not establish execution; exact selected effect terminal mandatory.', [68, 273, 58], ['membership'])
    task('membership', 'Validate only exact Paladin spell-list membership',
         'PDF55–56. Exactly38 named spells in5 level groups (13/11/6/4/4), school and C/R/M markers. Membership does not execute any spell. Positive: Divine Smite level1 belongs; negative: Cleric-only Sacred Flame is not a normal Paladin-list spell; Blessed Warrior and Devotion use separate explicit feature exceptions. #503 remains four-class only.', [194])
    spell_list={1:['Bless','Command','Cure Wounds','Detect Evil and Good','Detect Magic','Detect Poison and Disease','Divine Favor','Divine Smite','Heroism','Protection from Evil and Good','Purify Food and Drink','Searing Smite','Shield of Faith'],2:['Aid','Find Steed','Gentle Repose','Lesser Restoration','Locate Object','Magic Weapon','Prayer of Healing','Protection from Poison','Shining Smite','Warding Bond','Zone of Truth'],3:['Create Food and Water','Daylight','Dispel Magic','Magic Circle','Remove Curse','Revivify'],4:['Aura of Life','Banishment','Death Ward','Locate Creature'],5:['Dispel Evil and Good','Geas','Greater Restoration','Raise Dead']}
    tasks[-1]['fixture']+=' Exact finite membership by spell level: '+json.dumps(spell_list,sort_keys=True)+'.'
    task('preparation', 'Confirm Paladin prepared counts and one Long Rest replacement',
         'PDF53–54. Start exactly2 level1 Paladin spells (Heroism/Searing Smite recommendations only). Counts by class level:2,3,4,5,6,6,7,7,9,9,10,10,11,11,12,12,14,14,15,15. Choose additions only on count increase; spell level must have a Paladin slot. Completed Long Rest replaces at most ONE; always-prepared do not count and still count as Paladin spells. Positive: level5 six level1/2 spells; negative: level3 spell at5,combined multiclass slot expanding preparation,replace2 per rest.', [384, 428], ['membership'])
    task('slots', 'Track all five Paladin slot columns and completed-rest recovery',
         'PDF53–54/25. Consume exact twenty-row five-column values. Positive: Paladin1 two level1 slots,Paladin5 four level1/two level2; Paladin20 4/3/3/3/2. All expended slots recover only on completed Long Rest; multiclass ceil(Paladin/2) contribution, never add single-class slots atop combined table. Negative: Short Rest recharge,interrupted rest,double recovery,slot6.', [428], ['multiclass'])
    task('lay-on-hands', 'Resolve Lay On Hands healing and Poisoned removal',
         'PDF53. Paladin1+,pool5×Paladin level,all recover completed Long Rest. Bonus Action touch creature including SELF; spend up to remaining pool healing capped maxHP; optional Poisoned removal costs5 poolHP and those5 do not heal. Positive: pool10 spend3 healing +5 poison removal leaves2. Negative: remote target,Action timing,insufficient pool,client-forged cure,5 cost also healing. Canonical pool shared with Restoring Touch.', [428, 504, 727, 64, 67])
    task('mastery', 'Confirm two proficient Paladin weapon mastery kinds and replacement',
         'PDF54. Exactly2 proficient weapon kinds; Longsword/Javelin examples,Halberd/Flail replacements examples only. Completed Long Rest may change kinds. Positive: two distinct proficient choices; negative:3,duplicate,unproficient,rest start. Consume exact8 mastery effects489–496 only when selected weapon supplies that property.', [384, 428, 430] + list(range(489, 497)))
    task('style', 'Confirm Paladin Fighting Style or Blessed Warrior',
         'PDF54. Paladin2 choose Fighting Style feat OR Blessed Warrior. Recommendation not entitlement. Positive: Defense or Blessed Warrior branch; negative:both or no source-qualified feat. Feat effects391–394 remain owned there.', [384, 391, 392, 393, 394])
    task('blessed-warrior', 'Confirm Blessed Warrior two Cleric cantrips and replacement',
         'PDF54/38. Only if chosen atPaladin2. Exactly2 Cleric cantrips count as Paladin spells with CHA; Guidance/Sacred Flame recommendations only. On EACH Paladin level replaceONE with another Cleric cantrip. Positive: Guidance+Sacred Flame CHA; negative:Wizard-only Fire Bolt,3 cantrips,WIS routing,two replacements,non-Paladin level trigger. Exact seven Cleric-list effect owners required,not503 membership alone.', [503, 384], ['style', 'casting'])
    task('smite', 'Apply Paladins Smite preparation and one slot-free cast',
         'PDF54. Paladin2 always Divine Smite prepared; separate once-per-completed-Long-Rest slot-free cast. Does not waive Bonus Action trigger,components or target rules; can also spend qualified slots. Positive:one free cast then slotted cast; negative:second free cast,bypassed hit window,automatic damage without actual spell execution.', [428], ['casting', 'preparation'])
    task('channel', 'Track Paladin Channel Divinity capacity and recovery',
         'PDF53–55. Paladin3–10 two uses,11–20 three. Short Rest recoversONE expended; Long Rest ALL. One shared class pool; choose exactlyONE class effect per use; saveDC from Paladin casting. Positive:2spent→1 after Short Rest; negative:Cleric pool substituted,Short Rest full recharge,extra use before11.', [427, 428], ['casting'])
    task('divine-sense', 'Resolve Divine Sense known types and consecration presence',
         'PDF55. Paladin3 Bonus Action spend1 Channel Divinity;10minutes or Incapacitated. Within60ft know creature location AND type for Celestial/Fiend/Undead; detect consecrated/desecrated places/objects presence, Hallow reference does not cast Hallow. Positive:type+location within60; negative:Humanoid,distance61,expired,incapacitated,unsupported authoritative consecration fact. No model inference or unrelated hidden lore.', [487, 67], ['channel'])
    task('subclass', 'Confirm Oath of Devotion at Paladin three',
         'PDF55–56. Exactly one Paladin subclass at3,features at or below Paladin level; Devotion only declared subclass. Positive:3 unlocks its3 features; negative:2,second subclass,unlisted oath,legacy auto-penalty for breaking oath. Oath narrative is GM discussion,not automatic grant removal.', [384])
    task('improvement', 'Offer qualified Paladin feats at four eight twelve sixteen',
         'PDF55/86–88. ASI OR another feat for which qualified,not both. ASI +2 one/+1 two,max20; general feat total level4+,source-specific prerequisites and repeatability. Positive:one qualifying choice at each4/8/12/16; negative:duplicate grant,unqualified feat,cap21. Spell Recall requires Spellcasting; Magic Initiate alone insufficient. All17 exact effect owners mandatory for their selected alternatives.', [384, 389] + list(range(385, 402)))
    task('extra-attack', 'Apply Paladin two attacks on own-turn Attack action',
         'PDF55/25. Paladin5:twice instead of once whenever taking Attack action on OWN turn. Automatic benefit at triggering confirmed Attack action; no optional activation cost. Positive:two attacks in one action; negative:third,off-turn,reaction,Magic action,stacking another class Extra Attack.', [65, 779])
    task('steed', 'Apply Faithful Steed preparation and one slot-free cast',
         'PDF55. Paladin5 always Find Steed prepared,once slot-free per completed Long Rest,not free components/casting time/mount entitlement. Positive:canonical spell receipt creates qualified steed once; negative:duplicate free use,unqualified entity,foreign mount ownership. Consume actual Find Steed and183/184/185 world lifecycle terminals,not catalogue alone.', [428, 183, 184, 185], ['casting', 'preparation'])
    task('aura', 'Apply Aura of Protection with recipient choice and inactivity',
         'PDF55. Paladin6 passive self-origin10ft Emanation,inactive Incapacitated. Self/allies saves add max(1,CHAmod). Multiple Paladins do not stack: recipient choosesONE while in them. Positive:CHA-2 gives+1,CHA+3 gives+3; negative:distance11,enemy,incapacitated,+3+2 stacked. Automatic membership/revision changes recalculate; choice only where source requires it.', [487, 67, 66])
    task('abjure', 'Resolve Abjure Foes saves restriction and exact endings',
         'PDF55. Paladin9 Magic action,present Holy Symbol OR weapon,spend1 Channel Divinity. Up to max(1,CHAmod) visible creatures within60ft; WIS save or Frightened1minute/until ANY damage. While so Frightened target can ONLY ONE on its turn:move,Action,or Bonus Action. Positive:CHA+3 three targets; negative:four,61ft,unseen,no presented item,spent pool. Track source-specific turn restriction and expiry; damage ends only its owned instances.', [488, 67, 66], ['channel'])
    task('courage', 'Apply Aura of Courage Frightened immunity and suppression',
         'PDF55. Paladin10,self/allies within Protection aura immune Frightened; already Frightened ally entering has no effect WHILE there. Positive:enter suppresses,exit restores unexpired source; negative:delete original condition,enemy,outside,inactive aura. Passive automatic benefit,not optional activation.', [488], ['aura'])
    task('radiant', 'Apply Radiant Strikes to Melee weapon and Unarmed hits',
         'PDF55. Paladin11 every attack-roll HIT with Melee weapon OR Unarmed Strike extra1d8 Radiant; no once/turn restriction. Positive:two eligible hits each gain1d8; negative:ranged weapon,save-only effect,miss,wrong die. Respect critical dice/mitigation with exact67/65/779 contracts.', [61, 65, 67, 779])
    task('restoring', 'Resolve Restoring Touch six named conditions through shared endings',
         'PDF55. Paladin14 while using Lay On Hands may also remove ONE OR MORE Blinded,Charmed,Deafened,Frightened,Paralyzed,Stunned;5 poolHP EACH,none of cost heals. Positive:two conditions cost10 in same atomic receipt; negative:Restrained/Exhaustion,insufficient pool,double healing,blanket clear unrelated instances. #504 only supplies its three named endings; other exact trusted adapters remain required.', [504, 488, 727, 777], ['lay-on-hands'])
    task('expansion', 'Replace Aura of Protection radius at Paladin eighteen',
         'PDF55. Paladin18 Protection Emanation becomes30ft,not extra aura or added30. All dependent Devotion/Courage/Nimbus memberships consume canonical expanded/inactive aura. Positive:distance30 inside,31 outside; negative:40 radius or independent aura remaining active when source incapacitated.', [], ['aura'])
    task('boon', 'Offer Paladin nineteen Epic Boon or another qualified feat',
         'PDF55/87–88. Paladin19 Epic Boon OR other qualified feat; Truesight recommendation only. Total level19+ for boons; ability caps30 only where effect permits. Positive:qualified Truesight; negative:automatically grant recommended boon,total18,Spell Recall without actual Spellcasting. Preserve exact17 feat effects and383/384 selection ownership.', [384, 389] + list(range(385, 402)))
    task('oath-spells', 'Apply exact five Devotion always-prepared spell rows',
         'PDF56. Devotion3 Protection from Evil and Good/Shield of Faith;5 Aid/Zone of Truth;9 Beacon of Hope/Dispel Magic;13 Freedom of Movement/Guardian of Faith;17 Commune/Flame Strike. Cumulative thereafter; excluded from prepared count,Paladin CHA. Positive:all10 at17; negative:early row,Cleric WIS,spending regular preparation count,subclass absent. Exact10 effect terminals mandatory.', [351, 614, 692], ['subclass', 'preparation', 'casting'])
    task('sacred-weapon', 'Confirm Sacred Weapon imbue attack bonus damage choice and light',
         'PDF56. Devotion3 when taking Attack action optionally spend1 Channel Divinity on ONE held Melee weapon.10minutes or reuse/no longer carrying/voluntary no-action end. Add max(1,CHAmod) to its attacks; EACH HIT choose normal type OR Radiant,not both. Bright20ft+Dim20ft further. Positive:CHA-1 yields+1,normal/Radiant independently; negative:ranged/unheld,second simultaneous weapon,damage bonus instead of attack,both types,extra activation action. Preserve no extra optional gate for already-active bonus.', [65, 66], ['subclass', 'channel'])
    task('devotion-aura', 'Apply Aura of Devotion Charmed immunity and suppression',
         'PDF56. Devotion7 self/allies in canonical Protection aura immune Charmed; Charmed ally entering condition has no effect WHILE there. Positive:enter suppresses,exit resumes unexpired source. Negative:permanent cure,enemy,outside,inactive aura,wrong level. Passive source-automatic membership,not optional activation.', [488], ['subclass', 'aura'])
    task('smite-protection', 'Apply Smite of Protection until next own-turn start',
         'PDF57. Devotion15 WHENEVER cast Divine Smite,self/allies have Half Cover while in canonical aura until start YOUR next turn. Automatic trigger,no extra optional action/resource. Positive:slot-free and slotted Divine Smite both trigger; negative:ordinary hit,other smite,expired,next enemy turn substituted,inactive aura. Cover +2AC/+2DEXsave only per179 and cover combination rules.', [66], ['subclass', 'smite', 'aura'])
    task('nimbus-use', 'Confirm Holy Nimbus use restoration and exact endings',
         'PDF57. Devotion20 Bonus Action imbue aura10minutes or voluntary no-action end.1use per completed Long Rest; expendONE level5 slot to restore use,no action. Positive:spent use restored with5slot; negative:4slot,free recharge,rest start,double spend. Inactive Protection aura suspends its benefits; no fabricated source-ending rule.', [428], ['subclass', 'slots', 'aura'])
    task('nimbus-ward', 'Apply Holy Ward only to Fiend or Undead forced saves',
         'PDF57. Active Holy Nimbus SELF Advantage on saves forced by Fiend OR Undead. Positive:known authoritative Fiend save; negative:Humanoid,ally benefit,ability check,expired,inactive aura,invented type from narration. Automatic benefit,not optional activation.', [61, 67], ['nimbus-use'])
    task('nimbus-damage', 'Apply Holy Nimbus enemy turn damage and sunlight',
         'PDF57. Active aura Bright Light that IS sunlight; whenever ENEMY starts ITS turn inside,CHAmod+PB Radiant damage,no dice/no invented minimum. Positive:CHA+5/PB6 deals11; negative:ally,entry midturn,duplicate turn event,expired/inactive aura. Consume authoritative enemy/membership facts and mitigation; no source-free hidden-state disclosure.', [66, 67], ['nimbus-use'])
    by = {r['id']: r for r in records}
    resources = {r['id'].rsplit('/', 1)[-1]: r for r in records if r['kind'] == 'resource' and '/oath-of-devotion/' not in r['id']}
    levels = []
    feature_tasks = {'spellcasting': ['casting','preparation','slots','membership'], 'paladins-smite':['smite'], 'channel-divinity':['channel','divine-sense'], 'paladin-subclass':['subclass','oath-spells','sacred-weapon'], 'ability-score-improvement':['improvement'], 'faithful-steed':['steed'], 'aura-of-protection':['aura'], 'abjure-foes':['abjure'], 'aura-of-courage':['courage'], 'radiant-strikes':['radiant'], 'restoring-touch':['restoring'], 'aura-expansion':['expansion'], 'epic-boon':['boon'], 'fighting-style':['style','blessed-warrior'], 'extra-attack':['extra-attack'], 'weapon-mastery':['mastery'], 'lay-on-hands':['lay-on-hands']}
    for n in range(1,21):
        row = next(r for r in records if r['kind']=='level' and r['level']==n)
        values = {name: next((v['value'] for v in r.get('values_by_level',[]) if v['level']==n),0) for name,r in resources.items()}
        grants = [by[k]['name'] for k in row['progression']['feature_refs']]
        deps = ['hp','slots','preparation','mastery']
        if n==1: deps += ['traits','equipment','lay-on-hands','casting','membership']
        else: deps += [f'level-{n-1}']
        for k in row['progression']['feature_refs']: deps += feature_tasks.get(k.rsplit('/',1)[-1],[])
        deps += {7:['devotion-aura'],15:['smite-protection'],20:['nimbus-use','nimbus-ward','nimbus-damage']}.get(n,[])
        if n in [5,9,13,17]: deps += ['oath-spells']
        task(f'level-{n}', 'Confirm Paladin first-character level one' if n==1 else f'Confirm Paladin level{n-1} to{n} application',
             f'PDF53 row{n}; PDF23 global advancement. Authoritative Paladin class-level {n-1}→{n}, one adjacent transition only; first-class1 uses creation HP. XP and live PB select canonical advancement by authoritative TOTAL CHARACTER level after transition, independently of Paladin class level. The progression.proficiency_bonus below is class-table source metadata ONLY and never a multiclass XP/PB eligibility operand. Unequal witness: Fighter4/Paladin1→Paladin2 reaches total6,requires14000XP/PB+3,not300XP/PB+2; Paladin2 features/slots/HP still use Paladin2. Exact Paladin row: '+json.dumps(row['progression'],sort_keys=True)+'; resources: '+json.dumps(values,sort_keys=True)+'. Positive: apply these grants/counts exactly once after all consumed features qualify. Negative: class-indexed XP/PB substitution,insufficient total-level XP,level skip,duplicate grant,foreign class,state from another source version,wrong count/slot cell. Repeated ASI at8/12/16 consumes original feature; passive benefits automatic. No grants from planning closure.', [], sorted(set(deps)))
        levels.append(dict(level=n,progression=row['progression'],values=values,grants=grants))
    task('later-1','Confirm later-class Paladin first-level application',
         'PDF24–25/53. Later-class Paladin1 requires confirmed multiclass eligibility; later-level max(1,d10+CON) or max(1,6+CON) HP,not10+CON. Apply only reduced martial/light/medium/shield traits and level1 features. XP and live PB select canonical advancement by TOTAL CHARACTER level after transition independently of Paladin1. Positive:Fighter4→Fighter4/Paladin1 reaches total5,requires6500XP/PB+3,not0XP/PB+2; STR13/CHA13 and existing primary threshold required. Negative:class-indexed XP/PB,missing current threshold,starting equipment/saves/skills/Heavy armor,repeated admission,wrong combined slots. All20 higher Paladin applications retain total-level versus class-level distinction.', [], ['multiclass','hp','casting','preparation','slots','mastery','lay-on-hands'])
    common = ('Every source predicate requires deterministic positive and independently violated negative fixtures. Authoritative preview is inert: no writes/dice; explicit confirm revalidates actor/campaign/source/class level/expected revision, targets, range, action, resource and costs. For class applications only,class and total levels must be integers1–20,not booleans or fractional values; class level cannot exceed total. Independently reject total21,unqualified current-class configurations and unknown source paths. Class-application XP/live PB select canonical TOTAL-level advancement independently of Paladin-level feature/slot/HP rows. Shared effect execution adds no Paladin or other-class admission requirement to its targets. Commit choices, injectable dice, derived state and history atomically. Durable receipt binds payload: retry/replay/reload/reconnect returns exact result; conflicting payload and stale state reject. Concurrent confirmations commit once. Inject cost/effect/history failures and prove complete isolated PostgreSQL rollback. Web AND trusted TUI expose availability/reasons, accessible confirmation/cancel/reconnect; changed web requires desktop/phone journeys, build and focused lint. For passive/source-automatic effects confirmation belongs to triggering action/turn, never invent optional activation. Unknown facts remain unavailable; no hidden target/HP/DC disclosure. Feature tests use isolated entitlements; creation and later-level applications consume separately qualified execution. Planning closure never qualifies implementation.')
    evidence = {
        'internal/rules/srd521/creation.go':'NewSupportedBuild admits fixed Human Soldier Fighter/Human Acolyte Cleric only; no Paladin creation.',
        'internal/rules/srd521/character.go':'Project rejects classes other than fighter/cleric and levels except their fixed1/Fighter2 envelope; no Paladin derived sheet.',
        'internal/rules/srd521/catalogue.go':'Catalogue records and ValidateReadiness retain unimplemented mechanics; importing Paladin/Devotion is metadata only.',
        'internal/game/advancement_preview.go':'prepareAdvancement is specifically Fighter1→2 at300XP and fixed/roll; no Paladin adapter.',
        'internal/game/spell_preview.go':'Cure Wounds/Shield of Faith fixed Cleric seams; no Paladin source eligibility.',
        'internal/engine/action_confirmation.go':'Durable confirmation/transaction seams exist; no Paladin feature handlers or source eligibility inferred.',
        'internal/handlers/character_rules.go':'Shared HTTP sheet route; projection guard remains authority,not generic class-name presence.',
        'frontend/src/components/start/CharacterGuidedForm.tsx':'Paladin class label appears in form; not source-qualified admission.',
        'frontend/src/lib/startupWorkflow.ts':'Oathbound Emblem is narrative starter data,not Paladin source equipment or Holy Symbol entitlement.',
        'frontend/src/components/character/ActionPreviewPanel.tsx':'Shared preview/confirmation UI seam; no Paladin execution evidence.',
        'tui/app.go':'Existing trusted action command seam is fixed supported fixtures; no Paladin feature route qualification.'}
    def owners(r):
        if r['kind']=='level': return [f"level-{r['level']}"]
        ident=r['id']
        if '/oath-of-devotion' in ident:
            for feature, targets in [('oath-of-devotion-spells',['oath-spells']),('sacred-weapon',['sacred-weapon']),('aura-of-devotion',['devotion-aura']),('smite-of-protection',['smite-protection']),('holy-nimbus',['nimbus-use','nimbus-ward','nimbus-damage'])]:
                if feature in ident: return targets
            return ['subclass']
        if r['kind']=='class': return ['hp','traits','multiclass']
        slug=ident.rsplit('/',1)[-1]
        for name, targets in feature_tasks.items():
            if slug==name: return targets
        for fragment,targets in [('starting-equipment',['equipment']),('skills',['traits']),('spell-slot',['slots']),('prepared-spells',['preparation']),('blessed-warrior',['blessed-warrior']),('paladins-smite',['smite']),('faithful-steed',['steed']),('channel-divinity-effect',['channel']),('abjure-channel',['abjure']),('courage-aura',['courage']),('restoring',['restoring']),('expanded-aura',['expansion']),('mastery',['mastery'])]:
            if fragment in slug:return targets
        if r.get('parent_id') in by:return owners(by[r['parent_id']])
        raise AssertionError(ident)
    matrix = [dict(record=r, execution_owners=owners(r), state='metadata only; Paladin/Devotion unqualified', engine=list(evidence)[:7], web=list(evidence)[7:10], tui=['tui/app.go']) for r in records]
    def key_owners(r,key):
        targets=owners(r)
        if key=='srd-5.2.1:mechanic/paladin/spell-list/p55-56/level-1-plus':return ['membership']
        if r['kind']=='class':
            if 'hit-die' in key:return ['hp']
            if '/multiclass/' in key:return ['multiclass']
            return ['traits']
        if r['id'].endswith('/feature/paladin/channel-divinity'):
            return ['divine-sense'] if 'divine-sense' in key else ['channel']
        if r['id'].endswith('/feature/paladin/spellcasting'):
            if any(s in key for s in ['charisma','holy-symbol']):return ['casting']
            return ['preparation']
        if r['id'].endswith('/feature/paladin/paladin-subclass'):return ['subclass']
        if r['id'].endswith('/feature/paladin/fighting-style'):return ['style']
        if 'holy-nimbus' in r['id']:
            if '/holy-ward/' in key:return ['nimbus-ward']
            if '/radiant-damage/' in key or '/sunlight/' in key:return ['nimbus-damage']
            return ['nimbus-use']
        return targets
    for x in matrix:
        x['key_owners']={k:key_owners(x['record'],k) for k in x['record']['required_mechanics']+x['record']['option_keys']}
    for t in tasks:
        t['source_records']=[x['record']['id'] for x in matrix if t['key'] in x['execution_owners'] or any(t['key'] in v for v in x['key_owners'].values())]
        t['source_keys']=sorted({k for x in matrix for k,v in x['key_owners'].items() if t['key'] in v})
        t['numeric_context']=[dict(record=x['record']['id'], **{k:v for k,v in x['record'].items() if k in ['parameters','options','values_by_level','requirement','always_prepared','progression','capacity','select']}) for x in matrix if t['key'] in x['execution_owners']]
    manifest = dict(schema_version=1, decision=327, base='c577e3e7a8dac94d579b65f58af444fe427a412e',source_sha256=PIN,
                    publication='PENDING shared reservations and accepted post-Monk baseline',
                    catalogues=catalogues,records=matrix,levels=levels,issues=tasks,shared_issues=shared_tasks(),common_acceptance=common,
                    spell_lists=dict(paladin=spell_list,cleric_cantrips=['Guidance','Light','Mending','Resistance','Sacred Flame','Spare the Dying','Thaumaturgy'],devotion={3:['Protection from Evil and Good','Shield of Faith'],5:['Aid','Zone of Truth'],9:['Beacon of Hope','Dispel Magic'],13:['Freedom of Movement','Guardian of Faith'],17:['Commune','Flame Strike']}),
                    evidence={p:dict(finding=f,sha256=hashlib.sha256((ROOT/p).read_bytes()).hexdigest()) for p,f in evidence.items()},
                    routing=dict(reserved_429_consumers=['hp'], prohibited_direct_429_consumers=['qualification','level-1..20','later-1'],
                                 requirement='Every flattened leaf must have mandatory dependency path through bounded native gates to337 after327 closure; dependency and containment DAGs checked separately.'),
                    supplements=dict(status='PENDING root accepted post-Monk prefix; no captured prefix is accepted',targets=[47,54,283,337]))
    # Rebuilding the source audit must not discard published IDs, native receipts or seals.
    manifest.update({k:v for k,v in previous.items() if k not in manifest})
    if previous.get('published_issues'):
        manifest['publication']=previous['publication']
        manifest['supplements']=previous['supplements']
    destination.write_text(json.dumps(manifest,indent=2)+'\n')
    return manifest


if __name__ == '__main__':
    m=build()
    print(f"Built unpublished audit: {len(m['records'])} records,20 rows,{len(m['issues'])} class terminals; no tracker writes")

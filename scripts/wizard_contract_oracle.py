"""Independent source contract literals for #332; no runtime admission."""
COMMON = """Prepare is inert, without dice or writes. Authoritative confirmation revalidates actor, campaign, source version, entitlement, revision, targets, action, timing and costs. Persist injected dice, choices, resources, derived sheet and public history atomically in isolated PostgreSQL. Same-ID retry/reload/reconnect returns the saved receipt; conflicting payload, stale and foreign state reject. Race two confirmations: one commit. Inject cost/effect/history failure: full rollback. Shared engine availability and unavailable reasons agree on web and trusted TUI; test confirmation, cancellation and reconnect, desktop/phone behavior, build and changed-file lint. Passive effects belong to their triggering confirmed event. No private HP, targets or lore disclosure. Unknown execution stays unavailable. No production admission, inference, deployment or Full Rules promotion. CC-BY-4.0 attribution and adaptation: docs/srd-5.2.1-conformance.md."""
CONTRACTS = {
 'hp': 'PDF23/77: first character Wizard1 gets6+CON; later class entry and later Wizard levels add max(1,D6+CON) OR max(1,4+CON); add one D6 Hit Die, no healing. Positive CON2 fixed6; reject D10, first-character6 on later entry, reroll on retry. Consume968 qualification of429 transaction, never die implementation in968.',
 'core': 'PDF77 first-character only: Intelligence primary, Intelligence/Wisdom saves, simple weapons, no armor. Positive Wizard first character; reject later-entry save/weapon grants or armor training.',
 'skills': 'PDF77 choose2 distinct Arcana, History, Insight, Investigation, Medicine, Nature, Religion. Positive Arcana/Insight; reject3, duplicate, Acrobatics or later-entry skill grant.',
 'equipment': 'PDF77 choose A:2 Daggers, Arcane Focus Quarterstaff, Robe, Spellbook, Scholar pack,5GP OR B:55GP. Positive exact A or B; reject both,54GP, missing book or automatic later-entry gear.',
 'later-entry': 'PDF23-26/77 require Intelligence13 and prerequisites of every existing class. Gain Hit Die and level1 features only; no core saves/skills/weapons/gear. Use actual multiclass slots rules separately. Positive Wizard1 total5; reject INT12, missing foreign-class prerequisites, first-character HP or core grants. Composite hit-die-and-level1-features key belongs here, requires hp.',
 'membership': 'PDF79-82 exact217 printed Wizard entries with source level/school/ritual/material metadata. No Phantasmal Force list eligibility from header151. Mind Spike listed80/header149 agrees Wizard; effect belongs331. Positive listed Fireball; reject header-only Phantasmal Force, Sorcerous Burst or list metadata treated as execution.',
 'casting': 'PDF77/104-106 Intelligence spellDC8+INT+liveTOTAL PB; attack INT+TOTAL PB; Arcane Focus OR own spellbook subject to real component rules. Actual effect and source-qualified mode required. Positive Wizard1 total5 PB3, INT3 DC14; reject Wizard PB2, WIS, arbitrary focus, unqualified effect. Consume separate component/slot/long-casting/stack adapters from330.',
 'cantrips': 'PDF77 initial3 Wizard cantrips;4 at Wizard4,5 at Wizard10. Long Rest may replace1 with another Wizard cantrip. Recommendations Light/Mage Hand/Ray of Frost optional. Positive Wizard4 total8 gets4; reject total4 Wizard1 gets4,2 replacements, non-list cantrip or unavailable effect.',
 'book-identity': 'PDF77 Tiny3lb spellbook100pages; appearance/material choice; only you or Identify reader. Contains known level1+ spells. Positive self reads; reject unrestricted third-party reading or cantrip-only book. Do not invent finite spell capacity from100pages.',
 'book-initial': 'PDF77 initial6 level1 Wizard spells; recommendations Detect Magic, Feather Fall, Mage Armor, Magic Missile, Sleep, Thunderwave optional. Positive6 distinct eligible choices; reject5,7, level2 or recommendation mandatory.',
 'book-learning': 'PDF77 each Wizard level after1 add2 Wizard spells whose level has slots on own Wizard table. Positive Wizard2 total6 adds2 level1; reject total6 level3 spell, extra3, combined multiclass slot eligibility.',
 'copy-found': 'PDF78 found Wizard level1+ spell, including scroll, must be preparable level. Decipher/transcribe costs2hours and50GP per spell level. Positive level2 costs4h100GP; reject3h,99GP, foreign list, cantrip, inaccessible level. Afterwards prepare like other book spells. Scroll item lifecycle retains existing item owner.',
 'copy-own': 'PDF78 copying own known book uses1hour10GP per spell level. Positive level2 costs2h20GP; reject1h,19GP or treating found foreign text as own.',
 'lost-book': 'PDF78 reconstruct only prepared spells using own-copy1hour10GP per level; remainder requires finding new spells. Positive prepared level2 reconstruct2h20GP; reject restoring all formerly known/unprepared spells free.',
 'preparation': 'PDF77 own Wizard table counts4/5/6/7/9/10/11/12/14/15/16/16/17/18/19/21/22/23/24/25, Wizard level1-20. Only level1+ own-book spells with own-table slot access. Always-prepared excluded from count yet Wizard spells. Positive Wizard2 total6 prepares5 level1; reject total6 count10 or combinedslots higher eligibility.',
 'prepare-rest': 'PDF77 finish Long Rest may change any prepared level1+ spells from own eligible book; add when table increases. Positive swap all4 at Wizard1; reject Short Rest generic swap, nonbook or wrong own-slot level.',
 'slots': 'PDF77 all nine slot columns from exact Wizard row; Long Rest restores all. Class level selects slots; PB/XP uses complete admitted canonical total level. Positive Wizard1 total5 has2 level1 slots PB3; reject total5 Wizard slots or bool/fractional/total21/class>total.',
 'multiclass-slots': 'PDF24-26 full Wizard levels contribute to combined Spellcaster table with exact rounding for each other class; preparation/learning independently as single-class Wizard; Pact slots separate yet cross-casting source rules retained. Positive Wizard2 total6 own learning level1 despite combined higher slots; reject total-level preparation or Pact slots merged into pool.',
 'ritual': 'PDF78/104-105 Ritual Adept may cast Ritual-tag spell from own book while reading it, need not prepared; normal casting time+10minutes, no slot, no upcast. Ordinary prepared Ritual-tag mode from104 needs no Ritual Adept. Positive unprepared inbook Detect Magic10minutes+Action; reject untagged, no book/read, instantaneous ritual, slot spend or ritual ignored casting Concentration/Magic action each turn.',
 'arcane-recovery': 'PDF78 finish Short Rest study book choose expended slots combined level<=ceil(WizardLevel/2), none6+. Once Long Rest. Positive Wizard5 budget3 picks2+1; reject total-based budget, slot6, unexpended slot, seconduse or no study.',
 'scholar': 'PDF78 Wizard2 choose one already proficient Arcana, History, Investigation, Medicine, Nature, Religion; gain Expertise. Positive proficient Arcana; reject Insight, unproficient, two skills or stacking duplicate expertise.',
 'subclass': 'PDF78/82 Wizard3 select canonical Evoker; grant only features of own Wizard level or lower, at3/6/10/14. Positive Wizard3 total7 Savant and Potent; reject Wizard2 total6, second subclass or metadata execution.',
 'asi': 'PDF79 actual feat admission at Wizard4/8/12/16: ASI OR another qualifying feat. Positive noncaster qualifying ordinary feat outside encounter; reject unqualified feat or universal actual-Spellcasting/active-encounter prerequisite. Only399 Boon of Spell Recall requires actual Spellcasting.',
 'epic-boon': 'PDF79 Wizard19 Epic Boon OR another qualifying feat; Spell Recall recommended only. Positive alternative qualified feat; reject recommendation mandatory, Wizard18 total19 or399 without actual Spellcasting.',
 'memorize': 'PDF79 Wizard5 finish Short Rest study book replace1 level1+ Wizard spell prepared for Spellcasting with another level1+ own book spell. Positive1 swap; reject2, cantrip, nonbook, always-prepared replacement or no study.',
 'mastery-select': 'PDF79 Wizard18 choose one level1 AND one level2 book spell, each Action casting time; always prepared outside normal count. Positive eligible1+2; reject reaction/BonusAction/ritual-long spell or samelevel pair.',
 'mastery-cast': 'PDF79 selected Mastery spells cast at will at lowest level without slot; higher level requires slot. Positive selected level2 at2 free; reject free upcast, unselected spell or arbitrary effect execution.',
 'mastery-replace': 'PDF79 finish Long Rest study book replace at most1 Mastery spell with eligible same-level Action book spell. Positive swap level1; reject2 swaps, level2-for-level1, Short Rest or no study.',
 'signature-select': 'PDF79 Wizard20 choose2 distinct level3 own-book spells, always prepared. Positive two eligible; reject duplicate, level2, nonbook or counting against prepared limit.',
 'signature-cast': 'PDF79 each selected Signature spell independently casts at level3 without slot once per Short OR Long Rest; higher-level casting spends slot. Positive firstfree leaves secondfree; reject shared counter, freeupcast, repeatfree before rest.',
 'signature-recovery': 'PDF79 Short OR Long Rest restores each independent Signature free-cast resource; never all slots on Short Rest. Positive spentfirst restores, second remains1; reject duplicate recharge, arbitrary elapsedtime recharge or one shared counter.',
 'evoker-identity': 'PDF82 canonical chosen Evoker Wizard3+, source selection and own level required. Positive Wizard3; reject total3 Wizard1 or another subclass.',
 'savant-initial': 'PDF82 at Wizard3 choose2 Wizard Evocation spells each<=level2 and add to book free. Positive2 eligible no gold/time copycost; reject3, level3, nonEvocation, nonWizard or automatic preparation.',
 'savant-new-slots': 'PDF82 whenever gain a NEW slot level in THIS CLASS add1 Wizard Evocation spell of level for which have own slots, free. Positive Wizard5 new3 choose eligible1 OR2 OR3; reject new combined slot only, same slot level upgrade,2 choices or nonEvocation.',
 'potent-attack': 'PDF82 damaging cantrip cast AT creature, attack MISS target takes half damage if any, NO additional effect. Positive Ray of Frost miss half/no speedreduction; reject full damage, hit altered, nondamaging or rider on miss.',
 'potent-save': 'PDF82 damaging cantrip cast AT creature, successful target save takes half damage if any, NO additional effect. Positive Acid Splash success half; reject full, no damage, extra rider or leveled spell.',
 'sculpt': 'PDF82 Wizard6 Evocation spell affecting OTHER creatures you can SEE; may choose<=1+spelllevel; chosen auto succeed saves, zero damage ONLY if normally half on success. Positive Fireball four visible others; reject5, self, unseen, nonEvocation, unconditional zero damage against different save result.',
 'empowered': 'PDF82 Wizard10 Wizard Evocation spell may add INT modifier to ONE damage roll of spell. Positive one roll+3; reject each missile/roll, WIS, nonWizard/nonEvocation or total10 Wizard1.',
 'overchannel-damage': 'PDF82 Wizard14 damaging Wizard spell WITH slot1-5 may maximize spell damage only on turn cast. Positive slot5 initial damage max; reject cantrip/freecast, slot6, later-turn damage or nonWizard.',
 'overchannel-reuse': 'PDF82 first Overchannel no adverse effect; reuse BEFORE finish Long Rest takes2d12 Necrotic PER SLOT LEVEL immediately AFTER casting, ignores Resistance/Immunity. Each further reuse adds1d12 per slotlevel; Long Rest resets. Positive second slot3=6d12, third slot3=9d12; reject resistance, pre-castdamage, Short Rest reset, fixed2d12 or duplicate replay charge.',
}

# Explicit nested-context ownership; each ID below includes every key/option in
# that record, except the multi-owner feature/class records listed separately.
CONTEXT_GROUPS = {
 'skills': ['choice/wizard/skills'],
 'equipment': ['choice/wizard/starting-equipment'],
 'cantrips': ['choice/wizard/cantrips'],
 'book-initial': ['choice/wizard/spellbook-initial'],
 'book-learning': ['choice/wizard/spellbook-level-gain','prerequisite/wizard/spellbook-level-gain'],
 'preparation': ['choice/wizard/prepared-spells','prerequisite/wizard/prepared-slot-level'],
 'copy-found': ['choice/wizard/copy-found-spell','prerequisite/wizard/copy-found-spell'],
 'copy-own': ['choice/wizard/copy-own-book','prerequisite/wizard/copy-own-book'],
 'lost-book': ['choice/wizard/replace-lost-book','prerequisite/wizard/replace-lost-book'],
 'ritual': ['feature/wizard/ritual-adept','prerequisite/wizard/ritual-eligibility'],
 'arcane-recovery': ['feature/wizard/arcane-recovery','prerequisite/wizard/arcane-recovery-budget','resource/wizard/arcane-recovery'],
 'scholar': ['feature/wizard/scholar','choice/wizard/scholar','prerequisite/wizard/scholar-proficiency'],
 'subclass': ['feature/wizard/wizard-subclass','choice/wizard/wizard-subclass'],
 'asi': ['feature/wizard/ability-score-improvement','choice/wizard/ability-score-improvement','prerequisite/wizard/ability-score-improvement-qualification'],
 'epic-boon': ['feature/wizard/epic-boon','choice/wizard/epic-boon','prerequisite/wizard/epic-boon-qualification'],
 'memorize': ['feature/wizard/memorize-spell','prerequisite/wizard/memorize-replacement'],
 'mastery-select': ['choice/wizard/mastery-level-1','choice/wizard/mastery-level-2','prerequisite/wizard/mastery-level-1','prerequisite/wizard/mastery-level-2'],
 'signature-select': ['choice/wizard/signature-spells','prerequisite/wizard/signature-eligibility'],
 'signature-cast': ['resource/wizard/signature-first','resource/wizard/signature-second','prerequisite/wizard/signature-first-binding','prerequisite/wizard/signature-second-binding'],
 'evoker-identity': ['subclass/wizard/evoker/identity','prerequisite/wizard/evoker/identity-level','prerequisite/wizard/evoker/identity-selection'],
 'savant-initial': ['feature/wizard/evoker/evocation-savant','prerequisite/wizard/evoker/evocation-savant-level','choice/wizard/evoker/savant-initial','prerequisite/wizard/evoker/savant-initial-book'],
 'savant-new-slots': ['choice/wizard/evoker/savant-new-slots','prerequisite/wizard/evoker/savant-new-slots-trigger'],
 'potent-attack': ['feature/wizard/evoker/potent-cantrip','prerequisite/wizard/evoker/potent-cantrip-level','choice/wizard/evoker/potent-trigger'],
 'sculpt': ['feature/wizard/evoker/sculpt-spells','prerequisite/wizard/evoker/sculpt-spells-level'],
 'empowered': ['feature/wizard/evoker/empowered-evocation','prerequisite/wizard/evoker/empowered-evocation-level','prerequisite/wizard/evoker/empowered-evocation-ability'],
 'overchannel-damage': ['prerequisite/wizard/evoker/overchannel-level','prerequisite/wizard/evoker/overchannel-slots'],
}
# Exact split fields of large records are reviewed individually, never inferred
# from a substring like hit-die (the multiclass composite is not HP-only).
CLASS_KEYS = {
 'primary-ability/intelligence':'core', 'saves/intelligence':'core',
 'saves/wisdom':'core', 'weapons/simple':'core', 'armor/none':'core',
 'level-1/all-core-traits-and-level-1-features':'core',
 'multiclass/hit-die-and-level-1-features':'later-entry',
 'multiclass/slots-from-character-creation':'multiclass-slots',
 'hit-die-sides':'hp',
}
SPELLCASTING_KEYS = {
 'spellcasting/wizard-spell-list':'membership',
 'spellcasting/intelligence':'casting',
 'spellcasting/arcane-focus-or-spellbook':'casting',
 'spellcasting/cantrip-replace-one-on-long-rest':'cantrips',
 'spellcasting/add-cantrip-at-levels-four-and-ten':'cantrips',
 'spellcasting/regain-all-slots-on-long-rest':'slots',
 'spellcasting/prepare-from-spellbook-with-available-slot-level':'preparation',
 'spellcasting/add-prepared-when-table-increases':'preparation',
 'spellcasting/always-prepared-excluded-from-count-but-wizard-spells':'preparation',
 'spellcasting/change-any-prepared-from-book-on-long-rest':'prepare-rest',
 'spellcasting/initial-cantrips':'cantrips', 'spellcasting/initial-prepared':'preparation',
 'spellcasting/cantrip-replacement-count':'cantrips',
 'spellbook/dimensions':'book-identity', 'spellbook/weight-pounds':'book-identity',
 'spellbook/pages':'book-identity',
}

def expected_context_owners(records):
    """Literal independent ownership oracle, including all nested prerequisites."""
    groups={i:o for o,ids in CONTEXT_GROUPS.items() for i in ids}
    answer={}
    prefix='srd-5.2.1:mechanic/wizard/'
    for r in records:
        ident=r['id'].split(':character/')[1]
        keys=set(r['required_mechanics']+r['option_keys'])
        if r.get('requirement'): keys.add(r['requirement']['key'])
        if ident=='class/wizard':
            owners={prefix+k:v for k,v in CLASS_KEYS.items()}
        elif ident=='feature/wizard/spellcasting':
            owners={prefix+k:v for k,v in SPELLCASTING_KEYS.items()}
        elif r['kind']=='level':
            owners={k:'level-'+str(r['level']) for k in keys}
        elif ident.startswith('prerequisite/wizard/table-'):
            owners={k:'slots' for k in keys}
        elif ident=='feature/wizard/spell-mastery':
            owners={prefix+'mastery/'+k:v for k,v in {
                'always-prepared':'mastery-select','at-will-lowest-level-no-slot':'mastery-cast',
                'higher-level-requires-slot':'mastery-cast',
                'long-rest-study-book-replace-one-with-eligible-same-level':'mastery-replace',
                'replacement-count':'mastery-replace'}.items()}
        elif ident=='feature/wizard/signature-spells':
            owners={prefix+'signature/'+k:v for k,v in {
                'always-prepared':'signature-select','each-selected-spell-separate-free-cast':'signature-cast',
                'higher-level-requires-slot':'signature-cast'}.items()}
        elif ident=='feature/wizard/evoker/overchannel':
            owners={prefix+'evoker/overchannel/'+k:v for k,v in {
                'when-cast-damaging-wizard-spell-with-qualifying-slot':'overchannel-damage',
                'may-deal-maximum-spell-damage-on-turn-you-cast-it':'overchannel-damage',
                'minimum-slot-level':'overchannel-damage','maximum-slot-level':'overchannel-damage',
                'first-use-no-adverse-effect':'overchannel-reuse',
                'reuse-before-finish-long-rest-necrotic-self-damage':'overchannel-reuse',
                'initial-reuse-dice-per-spell-slot-level':'overchannel-reuse',
                'immediately-after-casting':'overchannel-reuse',
                'ignores-resistance-and-immunity':'overchannel-reuse',
                'each-further-use-before-finish-long-rest-increases-damage-per-spell-level':'overchannel-reuse',
                'initial-reuse-dice':'overchannel-reuse','die-sides':'overchannel-reuse',
                'further-use-dice-increase':'overchannel-reuse'}.items()}
        else:
            assert ident in groups, ('unclassified context',ident)
            owners={k:groups[ident] for k in keys}
            if ident=='choice/wizard/evoker/potent-trigger':
                owners[prefix+'evoker/potent/target-succeeds-on-save-against-cantrip']='potent-save'
        assert owners.keys()==keys, ('complete contextual keys',ident)
        answer[r['id']]=owners
    return answer

def expected_level_fixture(level, canonical_row):
    import json
    return (f'PDF23–26/77–82 Wizard class{level}; exact source row '+json.dumps(canonical_row,sort_keys=True)+
            '. Require adjacent prior Wizard level and qualified whole admitted class configuration; XP/livePB select TOTAL canonical level. Class/total integers1–20, never bool/fraction; class<=total. Witness Wizard1 total5 PB3 slots2 and Wizard2 total6 PB3 slots3, preparation5, learned spells level1. Reject total21, unknown other class, skipped level, duplicate grant or using total level for Wizard HP/features/slots. First-character1 and later-class1 are separate. Grant all required feature alternatives atomically.')

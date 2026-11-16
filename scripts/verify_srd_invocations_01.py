#!/usr/bin/env python3
"""Offline pinned pp72–73 comparison to actual #302 metadata, without Go."""
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
PIN = '8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87'
C = 'srd-5.2.1:character/'
M = 'srd-5.2.1:mechanic/'
B = 'warlock/invocations/'
P = 'srd-5.2.1:planning/invocations/'


def source_records():
    out = []
    def add(kind, slug, name, source, page, parent, keys=(), nums=None):
        r = dict(id=C+kind+'/'+B+slug, kind=kind, name=name,
                 sources=[dict(inventory_id=P+source, page=page)], parent_id=parent,
                 level=0, choice_refs=[], resource_refs=[], prerequisite_refs=[],
                 required_mechanics=[M+B+k for k in keys], select=0, option_keys=[], capacity=0)
        if nums:
            r['parameters'] = {M+B+k:v for k,v in nums.items()}
            r['required_mechanics'] += list(r['parameters'])
        out.append(r)
        return r
    def pred(r, suffix, key, refs, nums=None):
        slug = r['id'].split('/')[-1]
        p = add('prerequisite',slug+'-'+suffix,r['name']+' '+suffix,
                r['sources'][0]['inventory_id'][len(P):],r['sources'][0]['page'],r['id'],[key])
        p['requirement'] = dict(key=M+B+key, record_refs=[C+x for x in refs],parameters=nums or {})
        r['prerequisite_refs'].append(p['id'])
    def feature(slug, name, page, level, keys, nums=None):
        r = add('feature',slug,name,slug,page,C+'choice/warlock/invocations',keys,nums)
        if level:
            pred(r,'level','class-level',['class/warlock'],{'minimum-level':level})
        return r
    def choice(r, suffix, keys, options):
        slug=r['id'].split('/')[-1]
        q=add('choice',slug+'-'+suffix,r['name']+' '+suffix,slug,r['sources'][0]['page'],r['id'],keys)
        q['select']=1
        q['option_keys']=[M+x for x in options]
        r['choice_refs'].append(q['id'])
        return q
    a=feature('agonizing-blast','Agonizing Blast',72,2,[
        'agonizing-blast/add-charisma-modifier-to-chosen-spell-damage-rolls',
        'agonizing-blast/repeatable-different-eligible-cantrip-each-time'])
    q=choice(a,'cantrip',['agonizing-blast/choose-known-warlock-cantrip-that-deals-damage'],['warlock/known-damaging-cantrips'])
    pred(q,'eligibility','known-damaging-warlock-cantrip',['choice/warlock/cantrips','feature/warlock/pact-magic'])
    for slug,name,page,level,spell in [
        ('armor-of-shadows','Armor of Shadows',72,0,'mage-armor'),
        ('ascendant-step','Ascendant Step',72,5,'levitate')]:
        r=feature(slug,name,page,level,[slug+'/cast-on-self-without-spell-slot'])
        r['required_mechanics'].append(M+'spells/'+spell)
        pred(r,'spellcasting','canonical-pact-magic-spellcasting',['feature/warlock/pact-magic'])
    feature('devils-sight','Devil’s Sight',72,2,[
        'devils-sight/see-normally-in-dim-light-and-darkness-magical-and-nonmagical'],{'devils-sight/range-feet':120})
    d=feature('devouring-blade','Devouring Blade',72,12,[
        'requires/thirsting-blade-invocation', 'devouring-blade/thirsting-blade-extra-attack-replacement'],
        {'devouring-blade/extra-attacks':2,'devouring-blade/replaced-extra-attacks':1})
    pred(d,'invocation','requires/thirsting-blade-invocation',['choice/warlock/invocations'])
    feature('eldritch-mind','Eldritch Mind',72,0,['eldritch-mind/advantage-on-constitution-saves-to-maintain-concentration'])
    s=feature('eldritch-smite','Eldritch Smite',72,5,[
        'requires/pact-of-the-blade-invocation','eldritch-smite/optional-on-hit-creature-with-your-pact-weapon',
        'eldritch-smite/expend-pact-magic-slot','eldritch-smite/extra-force-damage/base-plus-dice-per-slot-level',
        'eldritch-smite/optional-prone-only-if-target-huge-or-smaller'],
        {'eldritch-smite/maximum-uses-per-turn':1,'eldritch-smite/slot-cost':1,
         'eldritch-smite/base-dice':1,'eldritch-smite/dice-per-slot-level':1,'eldritch-smite/die-sides':8})
    pred(s,'invocation','requires/pact-of-the-blade-invocation',['choice/warlock/invocations'])
    pred(s,'slot','eldritch-smite/canonical-slot-cost-and-level',['resource/warlock/pact-slots','feature/warlock/pact-magic'],{'slots-required':1})
    choice(s,'prone',['eldritch-smite/choose-whether-to-apply-prone'],['conditions/prone','warlock/invocations/eldritch-smite/no-prone'])
    e=feature('eldritch-spear','Eldritch Spear',73,2,[
        'eldritch-spear/on-cast-increase-chosen-spell-range-by-feet-per-warlock-level',
        'eldritch-spear/repeatable-different-eligible-cantrip-each-time'],{'eldritch-spear/feet-per-warlock-level':30})
    pred(e,'formula','eldritch-spear/range-formula-warlock-level',['class/warlock'])
    q=choice(e,'cantrip',['eldritch-spear/choose-known-damaging-warlock-cantrip-with-range-at-least-ten-feet'],['warlock/known-damaging-cantrips-with-range-at-least-ten-feet'])
    pred(q,'eligibility','known-damaging-warlock-cantrip-minimum-range',['choice/warlock/cantrips','feature/warlock/pact-magic'],{'minimum-range-feet':10})
    f=feature('fiendish-vigor','Fiendish Vigor',73,2,[
        'fiendish-vigor/cast-false-life-on-self-without-spell-slot',
        'fiendish-vigor/temporary-hit-points-use-highest-number-on-die-without-rolling'])
    f['required_mechanics'].append(M+'spells/false-life')
    pred(f,'spellcasting','canonical-pact-magic-spellcasting',['feature/warlock/pact-magic'])
    g=feature('gaze-of-two-minds','Gaze of Two Minds',73,5,[
        'gaze-of-two-minds/bonus-action-touch-willing-creature-perceive-through-its-senses',
        'gaze-of-two-minds/duration-until-end-of-your-next-turn',
        'gaze-of-two-minds/subsequent-turn-bonus-action-maintain-only-on-same-plane',
        'gaze-of-two-minds/maintenance-extends-until-end-of-your-next-turn-otherwise-ends',
        'gaze-of-two-minds/benefit-from-target-special-senses',
        'gaze-of-two-minds/while-perceiving-cast-from-either-space-only-within-range'],{'gaze-of-two-minds/casting-range-feet':60})
    choice(g,'target',['gaze-of-two-minds/choose-willing-touched-creature'],['targets/willing-touched-creature'])
    choice(g,'casting-origin',['gaze-of-two-minds/choose-spell-origin-within-range'],['spaces/self','spaces/connected-creature'])
    pred(g,'spellcasting','canonical-pact-magic-spellcasting',['feature/warlock/pact-magic'])
    d=feature('gift-of-the-depths','Gift of the Depths',73,5,[
        'gift-of-the-depths/breathe-underwater','gift-of-the-depths/swim-speed-equals-your-speed',
        'gift-of-the-depths/cast-water-breathing-without-spell-slot',
        'gift-of-the-depths/recover-casting-on-finish-long-rest'],{'gift-of-the-depths/casting-uses':1})
    d['required_mechanics'].append(M+'spells/water-breathing')
    pred(d,'spellcasting','canonical-pact-magic-spellcasting',['feature/warlock/pact-magic'])
    p=feature('gift-of-the-protectors','Gift of the Protectors',73,9,[
        'requires/pact-of-the-tome-invocation','gift-of-the-protectors/new-page-in-book-of-shadows-when-conjured',
        'gift-of-the-protectors/with-your-permission-creature-takes-action-to-write-own-name',
        'gift-of-the-protectors/name-capacity-equals-charisma-modifier-minimum-one',
        'gift-of-the-protectors/named-creature-reduced-to-zero-hit-points-but-not-killed-outright',
        'gift-of-the-protectors/trigger-drops-creature-to-one-hit-point-instead',
        'gift-of-the-protectors/after-trigger-no-creature-benefits-until-you-finish-long-rest',
        'gift-of-the-protectors/magic-action-touch-page-to-erase-name'],
        {'gift-of-the-protectors/minimum-names':1,'gift-of-the-protectors/replacement-hit-points':1,'gift-of-the-protectors/shared-trigger-uses':1})
    pred(p,'invocation','requires/pact-of-the-tome-invocation',['choice/warlock/invocations'])
    choice(p,'names',['gift-of-the-protectors/choose-permitted-creature-names-up-to-symbolic-capacity'],['targets/permitted-creature-writing-own-name'])
    choice(p,'erase',['gift-of-the-protectors/choose-name-to-erase'],['targets/name-on-touched-page'])
    return out


def compare(records):
    expected=source_records()
    assert records == expected, 'pp72–73 semantic graph mismatch'


def verify(pdf):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest()==PIN, 'PDF pin'
    def column(page,right):
        raw=subprocess.check_output(['pdftotext','-f',str(page),'-l',str(page),'-layout','-x','305' if right else '0','-y','0','-W','310','-H','792',str(pdf),'-'],text=True)
        return ' '.join(re.sub(r'-\n\s*','',raw.replace('\u00ad','')).split())
    text=column(72,True)+' '+column(73,False)+' '+column(73,True)
    # These clauses bind the independent semantic transcription to the pinned edition.
    clauses=[
        'Level 2+ Warlock, a Warlock Cantrip That Deals Damage',
        'Choose one of your known Warlock cantrips that deals damage',
        'add your Charisma modifier to that spell’s damage rolls',
        'Each time you do so, choose a different eligible cantrip',
        'cast Mage Armor on yourself without expending a spell slot',
        'Level 5+ Warlock', 'cast Levitate on yourself without expending a spell slot',
        'see normally in Dim Light and Darkness— both magical and nonmagical—within 120 feet',
        'Level 12+ Warlock, Thirsting Blade Invocation', 'confers two extra attacks rather than one',
        'Advantage on Constitution saving throws that you make to maintain Concentration',
        'Level 5+ Warlock, Pact of the Blade Invocation',
        'Once per turn when you hit a creature with your pact weapon',
        'expend a Pact Magic spell slot to deal an extra 1d8 Force damage',
        'plus another 1d8 per level of the spell slot', 'Prone condition if it is Huge or smaller',
        'has a range of 10+ feet', '30 times your Warlock level',
        'cast False Life on yourself without expending a spell slot',
        'don’t roll the die for the Temporary Hit Points; you automatically get the highest number on the die',
        'Bonus Action to touch a willing creature', 'until the end of your next turn',
        'same plane of existence', 'Bonus Action on subsequent turns to maintain this connection',
        'connection ends if you don’t maintain it', 'benefit from any special senses',
        'cast spells as if you were in your space or the other creature’s space', 'within 60 feet',
        'breathe underwater', 'Swim Speed equal to your Speed',
        'cast Water Breathing once without expending a spell slot', 'finish a Long Rest',
        'Level 9+ Warlock, Pact of the Tome Invocation', 'new page appears in your Book of Shadows when you conjure it',
        'With your permission, a creature can take an action to write its name',
        'number of names equal to your Charisma modifier (minimum of one name)',
        'reduced to 0 Hit Points but not killed outright', 'drops to 1 Hit Point instead',
        'no creature can benefit from it until you finish a Long Rest',
        'As a Magic action, you can erase a name on the page by touching it']
    for clause in clauses:
        assert clause in text, 'missing PDF clause: '+clause
    actual=json.loads((ROOT/'internal/rules/srd521/testdata/character-invocations-01.json').read_text())
    assert actual['schema_version']==1 and actual['source']['sha256']==PIN
    assert actual['source']['license']=='CC-BY-4.0' and actual['source']['attribution'] and actual['source']['changes']
    compare(actual['records'])
    inventory=json.loads((ROOT/'docs/investigations/issue-74-character-inventory.json').read_text())
    expected={r['id']:r['source_page'] for r in inventory['records'] if r['batch_key']=='options-invocations-01'}
    identities=[r for r in actual['records'] if r['kind']=='feature']
    assert len(identities)==len(expected)==12
    assert {r['sources'][0]['inventory_id']:r['sources'][0]['page'] for r in identities}==expected
    print(f'12 invocation identities and {len(actual["records"])} complete semantic records match pinned pp72–73')


if __name__=='__main__':
    if len(sys.argv)==3 and sys.argv[1]=='--controls':
        compare(json.loads((ROOT/'internal/rules/srd521/testdata/character-invocations-01.json').read_text())['records'])
        controls=json.loads(Path(sys.argv[2]).read_text())
        assert controls
        for records in controls:
            try:
                compare(records)
            except AssertionError:
                continue
            raise AssertionError('contrary graph accepted')
        print(f'{len(controls)} contrary graphs rejected')
    elif len(sys.argv)==2:
        verify(Path(sys.argv[1]))
    else:
        raise SystemExit('usage: verify_srd_invocations_01.py PDF | --controls JSON')

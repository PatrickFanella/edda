#!/usr/bin/env python3
"""Direct pinned source comparison for complete consumed spell/item contracts and domains."""
import hashlib
import json
from pathlib import Path
import re
import subprocess
from verify_paladin_consumed_sources import source,CASTING_CLAUSES
ROOT=Path(__file__).resolve().parents[1]
PATH=ROOT/'docs/investigations/issue-329-consumed-source-contracts.json'
PIN='8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87'
ITEMS={'Cloak of Protection':(216,'Cloak of the Bat'),'Gauntlets of Ogre Power':(223,'Gem of Brightness'),'Ring of Jumping':(238,'Ring of Mind Shielding'),'Ring of Protection':(238,'Ring of Regeneration'),'Spell Scroll':(244,'Copying a Scroll into a Spellbook.'),'Wand of Magic Detection':(251,'Wand of Magic Missiles')}

def extract(pdf):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest()==PIN
    spells=source(pdf)
    items={}
    for name,(page,end) in ITEMS.items():
        raw=subprocess.check_output(['rtk','proxy','pdftotext','-f',str(page),'-l',str(page),'-raw',str(pdf),'-'],text=True)
        # Exact heading line,not a reference to an item elsewhere in the page.
        start=re.search('^'+re.escape(name)+'$',raw,re.M);assert start,name
        tail=raw[start.start():];stop=tail.index(end);block=tail[:stop].strip()
        block=re.sub(r'-\s*\n','',block)
        items[name]=dict(page=page,complete_contract=' '.join(block.split()))
    casting={}
    for n in [104,105,106,206]:
        raw=subprocess.check_output(['rtk','proxy','pdftotext','-f',str(n),'-l',str(n),'-raw',str(pdf),'-'],text=True)
        casting[str(n)]=' '.join(re.sub(r'-\s*\n','',raw).split())
    return dict(source_sha256=PIN,license='CC-BY-4.0',attribution=json.loads((ROOT/'internal/rules/srd521/testdata/character-rogue.json').read_text())['source']['attribution'],changes='Complete consumed source contracts transcribed with whitespace and PDF hyphenation normalized; source IDs, dispatch and qualification decisions are Edda additions. No execution imported.',spells={name:spells[name] for name in ['Light','Detect Magic','Aid','Jump']},items=items,casting=casting,canonical_spell_names=sorted(k for k in spells if not k.startswith('__')))

def verify(pdf,m):
    actual=extract(pdf);assert actual==json.loads(PATH.read_text()),'complete actual consumed source contract drift'
    names=sorted(x['record']['name'] for x in m['dispatch_inventory'] if x['record']['family']=='spells')
    assert names==actual['canonical_spell_names'] and len(names)==339,'complete canonical identity domain'
    scroll=actual['items']['Spell Scroll']['complete_contract']
    table=re.findall(r'(Cantrip|[1-9]) (Common|Uncommon|Rare|Very Rare|Legendary) (\d+) \+(\d+)',scroll)
    assert [(n,int(dc),int(atk)) for n,_,dc,atk in table]==[(n,dc,atk) for n,dc,atk in zip(['Cantrip']+list(map(str,range(1,10))),[13,13,13,15,15,17,17,18,18,19],[5,5,5,7,7,9,9,10,10,11])]
    for clause in ['normal casting time','If the casting is interrupted, the scroll isn’t lost','DC equals 10 plus the spell’s level','On a failed check, the spell disappears from the scroll with no other effect']:
        assert clause in scroll,clause
    wand=actual['items']['Wand of Magic Detection']['complete_contract']
    assert all(x in wand for x in ['3 charges','While holding it','1 charge','Detect Magic','1d3 expended charges daily at dawn'])
    assert 'Attunement' not in wand and 'destroy' not in wand
    for n in ['Cloak of Protection','Ring of Protection']:
        assert all(x in actual['items'][n]['complete_contract'] for x in ['Requires Attunement','+1 bonus to Armor Class and saving throws','wear'])
    assert 'Strength is 19 while you wear' in actual['items']['Gauntlets of Ogre Power']['complete_contract']
    assert '19 or higher without them' in actual['items']['Gauntlets of Ogre Power']['complete_contract']
    assert 'can target only yourself' in actual['items']['Ring of Jumping']['complete_contract']
    for clause in CASTING_CLAUSES:assert any(clause in text for text in actual['casting'].values()),clause
    assert 'doesn’t expend any of the user’s spell slots' in actual['casting']['206']
    assert 'the GM determines when the item recharges' in actual['casting']['206']
    assert 'once on each of its turns' in actual['spells']['Jump'].lower() and 'spending 10 feet of movement' in actual['spells']['Jump']
    return actual

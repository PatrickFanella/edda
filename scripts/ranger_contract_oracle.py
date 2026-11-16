"""Independent readable oracle for the audited Ranger spell and shared contracts.

The literals preserve the full bounded numeric, alternative and timing predicates.
Verification never derives expected contracts from candidate manifests or receipts.
The source verifier separately binds these scopes to the pinned PDF. Body seals
protect the complete published envelope in addition to the readable scope oracle.
"""
import hashlib

EXECUTION_PREREQUISITES = (61, 64, 65, 66, 67, 273, 353)
EXECUTION_CLASS_CONSUMERS = (
    'spellcasting', 'extra-attack', 'roving', 'relentless-hunter', 'precise-hunter',
    'feral-senses', 'foe-slayer', 'tireless-hp', 'slots', 'colossus-slayer',
    'horde-breaker', 'escape-the-horde', 'multiattack-defense',
    'superior-hunters-prey', 'superior-hunters-defense',
)
HUNTER_ALTERNATIVES = {
    'hunters-prey': ('colossus-slayer', 'horde-breaker'),
    'defensive-tactics': ('escape-the-horde', 'multiattack-defense'),
}


def validate_spell_contracts(candidate):
    actual = {x['key']: x for x in candidate['tasks'] if x.get('spell')}
    assert actual.keys() == SPELL_CONTRACTS.keys(), 'complete bounded spell roster'
    for key, expected in SPELL_CONTRACTS.items():
        contract = actual[key]
        assert contract['issue'] == expected['issue'], ('published spell identity', key)
        assert contract['prerequisites'] == expected['prerequisites'], ('spell output owners', key)
        assert contract['depends'] == expected['depends'], ('split spell execution dependencies', key)
        assert contract['spell'] == expected['spell'], ('spell identity', key)
        assert type(contract['source_page']) is int and contract['source_page'] == expected['page'], ('spell page', key)
        assert contract['fixture'] == expected['fixture'], ('source execution predicates', key)
        assert expected['fixture'] in contract['body'], ('published execution scope', key)
        assert hashlib.sha256(contract['body'].encode()).hexdigest() == expected['body_sha256'], ('complete spell body', key)


def validate_hunter_ownership(candidate, rows, path_to):
    ownership = {
        c['id'].removeprefix('srd-5.2.1:character/'): {
            k.removeprefix('srd-5.2.1:mechanic/ranger/'): owner for k, owner in c['key_owners'].items()
        }
        for c in candidate['record_matrix']
    }
    assert ownership == SOURCE_OWNERS, 'complete contextual execution ownership'
    alternatives = {x for values in HUNTER_ALTERNATIVES.values() for x in values}
    for context in candidate['record_matrix']:
        record = context['record']
        if '/ranger/hunter/' not in record['id']:
            continue
        slug = record['id'].rsplit('/', 1)[-1]
        if record['kind'] == 'subclass' or slug == 'identity-level':
            base = 'hunter-identity'
        elif record['kind'] == 'prerequisite' and slug not in alternatives:
            base = record['parent_id'].rsplit('/', 1)[-1]
        else:
            base = slug
        expected = {}
        for key in record['required_mechanics'] + record['option_keys']:
            option = key.rsplit('/', 1)[-1]
            expected[key] = option if record['kind'] == 'choice' and option in alternatives else base
        assert context['key_owners'] == expected, ('nested Hunter execution owner', record['id'])
    for selector, alternatives in HUNTER_ALTERNATIVES.items():
        assert set(alternatives) <= set(rows[selector]['depends']), ('nested execution alternatives', selector)
    for level, required in [(3, HUNTER_ALTERNATIVES['hunters-prey']), (7, HUNTER_ALTERNATIVES['defensive-tactics'])]:
        for owner in required:
            assert path_to(rows, 'level-' + str(level), owner), ('granting-level nested execution', level, owner)


def validate_protocols(candidate, rows):
    gate = rows['execution-dependencies']
    assert gate['issue'] == 978 and gate['parent'] == 328, 'finite protocol gate identity'
    assert gate['prerequisites'] == list(EXECUTION_PREREQUISITES), 'exact protocol prerequisite roster'
    assert gate['depends'] == [], 'protocol gate cannot depend on consumers'
    expected = set(EXECUTION_CLASS_CONSUMERS) | set(SPELL_CONTRACTS)
    actual = {key for key, row in rows.items() if 'execution-dependencies' in row['depends']}
    assert actual == expected, 'finite protocol consumer roster'
    assert all(str(n) in candidate['qualification_paths'] for n in EXECUTION_PREREQUISITES), 'mandatory protocol path declarations'


def validate_native_obligations(candidate, graph, rows):
    assert graph[978]['blocked_by'] == list(EXECUTION_PREREQUISITES), 'native protocol prerequisites'
    for key in set(EXECUTION_CLASS_CONSUMERS) | set(SPELL_CONTRACTS):
        assert 978 in graph[rows[key]['issue']]['blocked_by'], ('native protocol consumer', key)

    def reachable(root, target):
        todo = [root]
        seen = set()
        while todo:
            current = todo.pop()
            if current == target:
                return True
            if current not in seen:
                seen.add(current)
                todo.extend(graph[current]['blocked_by'])
        return False

    # These obligations are derived from the source roster, never from optional
    # path declarations that a candidate could delete along with an edge.
    assert len(candidate['tasks']) == 113 and len(candidate['gates']) == 9, 'finite implementation inventory'
    for contract in candidate['tasks']:
        assert reachable(337, contract['issue']), ('mandatory execution path', contract['key'])
    for key in SPELL_CONTRACTS:
        assert all(reachable(root, rows[key]['issue']) for root in [279, 280]), ('shared spell qualification', key)

SPELL_CONTRACTS = {'spell-alarm': {'spell': 'Alarm',
                 'page': 107,
                 'fixture': '1min orRitual/VSM/30ft/8h; door/window/area<=20ftCube; chosen exclusions; '
                            'creaturetouch/entry triggers audible10sec within60ft ORmental within1mile wakescaster. '
                            'Positive:excludedcreature noalarm;negative:21ftCube,mental beyondmile.',
                 'body_sha256': 'a3a82120f00a0c57d4e9e8327e600a2c96ad76b8b850049534b0bcc3d6c5d268',
                 'issue': 1035,
                 'prerequisites': [],
                 'depends': ['execution-dependencies']},
 'spell-animal-friendship': {'spell': 'Animal Friendship',
                             'page': 107,
                             'fixture': 'Action/VSM/30ft/24h; visibleBeast WISfailCharmed;caster ORallydamage '
                                        'ends;slotL targetsL. Positive:L2 twoBeasts;negative:Humanoid,success,no '
                                        'lineofsight. Consume488.',
                             'body_sha256': '6649744b22f4f3ed0037f68241b91b5f2f6e8a40986d07228995bd7f55e274d4',
                             'issue': 1036,
                             'prerequisites': [488],
                             'depends': ['execution-dependencies']},
 'spell-animal-messenger-admission': {'spell': 'Animal Messenger',
                                      'page': 108,
                                      'fixture': 'Action ORRitual/VSM/30ft; visibleTinyBeast,CHA save;CR!=0 '
                                                 'autosuccess; visiteddestination,generalrecipient,<=25words. '
                                                 'Positive:CR0fail;negative:Small,CR1,26words,unvisited.',
                                      'body_sha256': '09f65b405e7bf8502f68d7bdf072ce1438cd75c4c02c092bd04e95110234c603',
                                      'issue': 1037,
                                      'prerequisites': [],
                                      'depends': ['execution-dependencies']},
 'spell-animal-messenger-journey': {'spell': 'Animal Messenger',
                                    'page': 108,
                                    'fixture': 'Acceptedmessage travels24h +48h/slotabove2,25miles/day '
                                               'or50flying;recipientmatchesdescription '
                                               'mimickedmessage;expirybeforearrival '
                                               'losesmessage/returnstocastingplace. Persist '
                                               'trustedroute/time/arrival,neverinfer travel or recipient. No '
                                               'independentNPCcombatAI.',
                                    'body_sha256': 'dce5eb2d998db3ebc0361e8f50db76f364d355c4c5be04887a6f7a33a27a6caa',
                                    'issue': 1038,
                                    'prerequisites': [],
                                    'depends': ['execution-dependencies', 'spell-animal-messenger-admission']},
 'spell-barkskin': {'spell': 'Barkskin',
                    'page': 112,
                    'fixture': 'BonusAction/touchwilling/VSM/1h;ACfloor17 onlyiflower,not+17,NOConcentration. '
                               'Positive:AC12→17;negative:AC20reduced,unwilling.',
                    'body_sha256': '68a5f2c8231937b037c4a242c78fe899ba19f055f426f3781acb2c314eb1da92',
                    'issue': 1039,
                    'prerequisites': [],
                    'depends': ['execution-dependencies']},
 'spell-commune-with-nature-facts': {'spell': 'Commune with Nature',
                                     'page': 116,
                                     'fixture': '1min '
                                                'ORRitual/VS/self/instant;outdoors3miles,naturalunderground300ft,constructedsettlement '
                                                'NOeffect. Choose3: '
                                                'settlements,portals,ONECR10+Celestial/Elemental/Fey/Fiend/Undead '
                                                'GMchoice,mostprevalentPlant/Mineral/Beastkind,waterlocations. '
                                                'ExplicitGM/source-authorizedfacts;no inferred oromniscientanswers. '
                                                'Negative:4facts,CR9,constructedarea.',
                                     'body_sha256': '4ba2cde15bf9c16ba18751ab644477beab61c70fac4e57b04bf767494e6cf47b',
                                     'issue': 1040,
                                     'prerequisites': [],
                                     'depends': ['execution-dependencies']},
 'spell-conjure-animals-aura': {'spell': 'Conjure Animals',
                                'page': 117,
                                'fixture': 'Action/VS/60ft/Concentration10min;ONE Largeintangible spectralpack in '
                                           'visibleunoccupiedspace;formcosmeticonly;STRsaveAdvantage whilewithin5ft. '
                                           'Notcreatureentities/no statblocks.',
                                'body_sha256': '70ab1f42bb3a1ba7d4fa7360e0f28a99d58c905da9cc4161e938af5ea55480ee',
                                'issue': 1041,
                                'prerequisites': [],
                                'depends': ['execution-dependencies']},
 'spell-conjure-animals-movement': {'spell': 'Conjure Animals',
                                    'page': 117,
                                    'fixture': 'OnYOURturn whenyoumove,pack canmoveupto30ft '
                                               'tovisibleunoccupiedspace;atomic source-linked '
                                               'positioning,negative:otherturn,31ft,occupied,inferredpositions.',
                                    'body_sha256': 'e291bcbead355eaf45ae37775df27e66be039ee9b58c3d291e7ad9fa5cd23c83',
                                    'issue': 1042,
                                    'prerequisites': [],
                                    'depends': ['execution-dependencies', 'spell-conjure-animals-aura']},
 'spell-conjure-animals-damage': {'spell': 'Conjure Animals',
                                  'page': 117,
                                  'fixture': 'Packmoveswithin10ft ofvisiblecreature ORvisiblecreatureenterswithin10ft '
                                             'ORendsownturnthere;optionalDEXsave '
                                             'ONCEperturn,fail3d10Slashing,success0,+1d10/slotabove3. '
                                             'Not2014summon;negative:secondtrigger,sightmissing,successdamage.',
                                  'body_sha256': '226f9913a321782b77564c9ac9bb9ca9ce97c2c1f78a2d1b18e2f8557b0ff391',
                                  'issue': 1043,
                                  'prerequisites': [],
                                  'depends': ['execution-dependencies', 'spell-conjure-animals-aura']},
 'spell-conjure-woodland-beings-aura': {'spell': 'Conjure Woodland Beings',
                                        'page': 118,
                                        'fixture': 'Action/VS/self/Concentration10min;10ftEmanation;entersvisiblecreaturespace '
                                                   'ORvisiblecreatureentry/endturn optionalWISsave '
                                                   'onceperturn;5d8Forcefail/halfsuccess,+1d8/slotabove4;notcreatureentities.',
                                        'body_sha256': 'fdbcfb34667772e568450ec113940decc7345806357f8f36cbc5863b1ba6018c',
                                        'issue': 1044,
                                        'prerequisites': [],
                                        'depends': ['execution-dependencies']},
 'spell-conjure-woodland-beings-disengage': {'spell': 'Conjure Woodland Beings',
                                             'page': 118,
                                             'fixture': 'Whileownspellactive maytakeDisengageasBonusAction,consume781 '
                                                        'exactbehavior/actioneconomy;notfreeDisengage '
                                                        'oradditionalBonusAction.',
                                             'body_sha256': '9c5c0a8939c5731492fa1ca770b0939ed6f9d2c10f444fbdf9160c794e47f0cc',
                                             'issue': 1045,
                                             'prerequisites': [781],
                                             'depends': ['execution-dependencies',
                                                         'spell-conjure-woodland-beings-aura']},
 'spell-darkvision': {'spell': 'Darkvision',
                      'page': 122,
                      'fixture': 'Action/touchWILLING/VSM/8h;Darkvision150ft. '
                                 'Positive:150ft;negative:60ftsubstitution,unwilling,seeingthroughTotalCover. '
                                 'Source-specificsense doesnotgrant hiddenfacts.',
                      'body_sha256': '428c4dac6f134e021215484b7ae12af764a6f3c558d464a4305937fbb41bd280',
                      'issue': 1046,
                      'prerequisites': [],
                      'depends': ['execution-dependencies']},
 'spell-dominate-beast-charm': {'spell': 'Dominate Beast',
                                'page': 125,
                                'fixture': 'Action/VS/60ft/Concentration1min;visibleBeast '
                                           'WISsave,Advantageiffightingcaster/allies;failCharmed488,EVERYdamagerepeatsWIS '
                                           'endingonsuccess;slot5/6/7+ duration10min/1h/8h. '
                                           'Noforcedcontrolbeforefailedsave.',
                                'body_sha256': '47b809bc038c08f315c372ab71a14d321a4f409c8444550ca7a02f447150fd16',
                                'issue': 1047,
                                'prerequisites': [488],
                                'depends': ['execution-dependencies']},
 'spell-dominate-beast-control': {'spell': 'Dominate Beast',
                                  'page': 125,
                                  'fixture': 'AcceptedownCharm/sameplane telepathiclink;YOURturn commands '
                                             'noaction,targetbestobeys onownturn;completedorderwithoutnewdirection '
                                             'selfprotect/freebehavior. CommandReaction requires BOTHcasterReaction '
                                             'andtargetReaction. Explicitbounded authoritativecommands '
                                             'notmodelinferredNPCactions.',
                                  'body_sha256': '0a8caf3daeb8cfca7f4488cb32e9073729231d43b75e8effd3eed90a5acea27b',
                                  'issue': 1048,
                                  'prerequisites': [488],
                                  'depends': ['execution-dependencies', 'spell-dominate-beast-charm']},
 'spell-druidcraft-weather': {'spell': 'Druidcraft',
                              'page': 126,
                              'fixture': 'Action/VS/30ft: Tinyharmlesssensor forecastsOWNlocation '
                                         'next24h,persists1round. '
                                         'Trustedweatherfactinputonly;negative:inventedweather,damagingeffect,otherlocation.',
                              'body_sha256': '97457f0b7e1f519e4c05ebc81e5cf91e8fc539c1d2b2a4ebaa9238c57cc6c9da',
                              'issue': 1049,
                              'prerequisites': [],
                              'depends': ['execution-dependencies']},
 'spell-druidcraft-nature': {'spell': 'Druidcraft',
                             'page': 126,
                             'fixture': 'Action/VS/30ft: chooseBloom(singleflower/seedpod/leafbud) '
                                        'ORharmlesssensoryeffect<=5ftCube ORlight/snuffcandle/torch/campfire. '
                                        'Nobroadplantgrowth/weathercontrol;persistselectedeffectonce.',
                             'body_sha256': '6172e1e8cc4d99c9f601e598e73b1581e8481755ca69f1beb6c9074668233519',
                             'issue': 1050,
                             'prerequisites': [],
                             'depends': ['execution-dependencies', 'spell-druidcraft-weather']},
 'spell-elementalism-air-earth': {'spell': 'Elementalism',
                                  'page': 127,
                                  'fixture': 'Action/VS/30ft;Airbreeze5ftCube '
                                             'ripple/stirdust/rustle/closeUNHELDdoorsshutters;Earththin '
                                             'dust/sand5ftsquare ORonehandwrittenwordindirt. Nohelddoorforce/damage.',
                                  'body_sha256': '2e9fe707ffdb7cd8f2e66ae75e28d7220a7267c7a5a37f0b3a31c818dee81153',
                                  'issue': 1051,
                                  'prerequisites': [],
                                  'depends': ['execution-dependencies']},
 'spell-elementalism-fire-water': {'spell': 'Elementalism',
                                   'page': 127,
                                   'fixture': 'Action/VS/30ft;harmlesscoloredscentedembers/smoke5ftCube '
                                              'canlightcandles/torches/lamps,scent1min;Watermist5ftCube '
                                              'OR1cupcleanwateropencontainer/surface,evaporates1min. '
                                              'Nopermanentwater/firedamage.',
                                   'body_sha256': '06a244eb28f96d9606ed90c288b1dd1a643ad7725b1d17da0505668d062e8698',
                                   'issue': 1052,
                                   'prerequisites': [],
                                   'depends': ['execution-dependencies', 'spell-elementalism-air-earth']},
 'spell-elementalism-sculpt': {'spell': 'Elementalism',
                               'page': 127,
                               'fixture': 'Action/VS/30ft;ONE dirt/sand/fire/smoke/mist/water<=1ftCube crudeshape1h. '
                                          'Negative:2ftCube,stone/metal,finetoolcreation,permanence.',
                               'body_sha256': '6325b689d015750445c4bb95ecca620b66d26e86267813c2e513773acb9ea60d',
                               'issue': 1053,
                               'prerequisites': [],
                               'depends': ['execution-dependencies', 'spell-elementalism-air-earth']},
 'spell-enhance-ability': {'spell': 'Enhance Ability',
                           'page': 127,
                           'fixture': 'Action/touch/VSM/Concentration1h;chooseSTR/DEX/INT/WIS/CHA,NOTCON;AdvantageABILITYCHECKSonly;slotL '
                                      'targetsL-1,abilitymaydiffer. '
                                      'Negative:saves/attacks,CONchoice,forcedwillingpredicate.',
                           'body_sha256': 'db13479b3bbb496066a072f73086043813cd873507f381c419cba6867a75e129',
                           'issue': 1054,
                           'prerequisites': [],
                           'depends': ['execution-dependencies']},
 'spell-ensnaring-strike-hit': {'spell': 'Ensnaring Strike',
                                'page': 128,
                                'fixture': 'BonusAction IMMEDIATELYWEAPONhit,V,self,Concentration1min;STRsave '
                                           'target,Large+Advantage;failRestrained727,successends. '
                                           'Negative:Unarmedonly,miss,rangedbaninvented,delayedwindow.',
                                'body_sha256': '7ad5b5f3ffff69666997145376807aee6a07fa71089d3b2973abf49d5d8fbdc3',
                                'issue': 1055,
                                'prerequisites': [727],
                                'depends': ['execution-dependencies']},
 'spell-ensnaring-strike-tick-escape': {'spell': 'Ensnaring Strike',
                                        'page': 128,
                                        'fixture': 'SAMEsourceRestrainedtargetstartturn1d6Piercing '
                                                   '+1d6/slotabove1;target ORcreaturewithinreach ACTIONSTRAthletics vs '
                                                   'spellDC successfulend. '
                                                   'Negative:freeescape,wrongcasterDC,duplicateevent,unrelatedRestrainedcleared.',
                                        'body_sha256': '9dd287674193caa348baf9dc44463bf3127f16949ab3017758fa0144630808de',
                                        'issue': 1056,
                                        'prerequisites': [727],
                                        'depends': ['execution-dependencies', 'spell-ensnaring-strike-hit']},
 'spell-entangle': {'spell': 'Entangle',
                    'page': 128,
                    'fixture': 'Action/VS/90ft/Concentration1min;20ftgroundsquare '
                               'DifficultTerrain;initialcreaturesEXCEPTcasterSTRsave,failRestrained727;ActionSTRAthletics '
                               'vsDCfreeitself. Terrain/sourceplantsendwithspell. '
                               'Noentryrestrain,automaticinitialcasterrestrain,otherrestraintsremoved.',
                    'body_sha256': '05b97c0b913db9f3da991db536af2b816ce009d7a4b2685d3b1ea4647a0c19b0',
                    'issue': 1057,
                    'prerequisites': [727],
                    'depends': ['execution-dependencies']},
 'spell-find-traps': {'spell': 'Find Traps',
                      'page': 131,
                      'fixture': 'Action/VS/120ft/instant;lineofsight '
                                 'createddangerobject/mechanismincludingAlarm/Glyph,NOTnaturalweakfloor/ceiling/sinkhole. '
                                 'Disclosepresence+generalDANGER,NEVERlocation. '
                                 'Trustedauthorizedtrapfacts;negative:exactcoordinates/naturalhazard/modelguesses.',
                      'body_sha256': '5947aec27a71bbd4f13d9e6facbce32572c9acf51e09832ddf8b74fd9d1610e3',
                      'issue': 1058,
                      'prerequisites': [],
                      'depends': ['execution-dependencies']},
 'spell-goodberry-create': {'spell': 'Goodberry',
                            'page': 137,
                            'fixture': 'Action/VSM/self;createTENindividualsource-linkedberries,24huneatenexpiry. '
                                       'Negative:11berries,expiryresetontransfer,duplicatecreation;authoritativeitems,notinventorylabels.',
                            'body_sha256': '39732db1be02428f9117228dda374a67d22820669cd1d06a4ee39f0c720d0d9c',
                            'issue': 1059,
                            'prerequisites': [],
                            'depends': ['execution-dependencies']},
 'spell-goodberry-eat': {'spell': 'Goodberry',
                         'page': 137,
                         'fixture': 'OnecreatureBonusAction '
                                    'eatONEunexpiredberry,consumeonce,heal1HP+sustain1day;source-owned berryidentity '
                                    'transfer/reload/retryatomic. '
                                    'NotAction,notWIShealing,notmultipleconsumerssameberry.',
                         'body_sha256': 'e48353d6164df6c951962d6b5f02bc8fba46e0a59cf1b9c730ae53af2e758d66',
                         'issue': 1060,
                         'prerequisites': [],
                         'depends': ['execution-dependencies', 'spell-goodberry-create']},
 'spell-gust-of-wind': {'spell': 'Gust of Wind',
                        'page': 138,
                        'fixture': 'Action/VSM/self/Concentration1min;60ftlong10ftwideLine;initial '
                                   'andendturnSTRsavefailpush15ftalongline away;movingcloser '
                                   'costs2ft/ft;laterYOURturnBonusActionredirect. Sourcegas/vapor '
                                   'disperse,unprotectedflameextinguish,protected50% savedroll. '
                                   'Negative:offaxis/failedpushdistance/repeatedendturn/initialredirectfree.',
                        'body_sha256': 'b3d276ee61f45908277a77951568dc174b162867f3249efe3efe20a34b5a0298',
                        'issue': 1061,
                        'prerequisites': [],
                        'depends': ['execution-dependencies']},
 'spell-hunters-mark-cast': {'spell': 'Hunter’s Mark',
                             'page': 141,
                             'fixture': 'BonusAction/V/90ft/Concentration1h;ONEvisiblecreature;EVERY attackrollhit '
                                        'extra1d6FORCE '
                                        '(notweapononly,notonceperturn);WISPerception/SurvivaltoFINDitAdvantage. '
                                        'Slot3-4duration8h,5+24h;noNecrotic/2014abilitychoice. '
                                        'CanonicaldieadapterFoeSlayerdistinct.',
                             'body_sha256': '314ba19f96f3488dc70f294a11276c7c08a5c74c3eb294a81d554de96366cb25',
                             'issue': 1062,
                             'prerequisites': [],
                             'depends': ['execution-dependencies']},
 'spell-hunters-mark-transfer': {'spell': 'Hunter’s Mark',
                                 'page': 141,
                                 'fixture': 'Onlyifcurrentmarkedtargetdrops0HP BEFOREspellends:BonusAction '
                                            'moveOWNmarktoNEWvisiblecreaturewithin90ft;SAMEcast/duration/concentration,NOnewslot/freeusecharge. '
                                            'Negative:alive,expired,foreignmark,transferanytime,private0HPleak.',
                                 'body_sha256': '571b6aa51f84e8327ddb6ae0b5cb6622c00c34faf1fef40af24774ee81a3190c',
                                 'issue': 1063,
                                 'prerequisites': [],
                                 'depends': ['execution-dependencies', 'spell-hunters-mark-cast']},
 'spell-jump': {'spell': 'Jump',
                'page': 143,
                'fixture': 'BonusAction/touchwilling/VSM/1min;ONCEoneachTARGETturn jumpupto30ft '
                           'spending10ftmovement;slotL targetsL. '
                           'Negative:tripledistanceoldvariant,freejump,unwilling,secondturnjump.',
                'body_sha256': '65a5a7a80ada17a8f45e0dc60d923b8e0fdf44734e84dfeb8ec257eba648fd4c',
                'issue': 1064,
                'prerequisites': [],
                'depends': ['execution-dependencies']},
 'spell-locate-animals-or-plants': {'spell': 'Locate Animals or Plants',
                                    'page': 144,
                                    'fixture': 'Action '
                                               'ORRitual/VSM/self/instant;namedKINDBeast/Plantcreature/nonmagicalplant,closestwithin5miles '
                                               'directionANDdistance,noneifabsent. Trustedfacts '
                                               'only;negative:specificindividual,6miles,hiddenalllocations,madeupresult.',
                                    'body_sha256': '880dbb099a682e0b26948f571fd1cf3290417fa2140b09c991ab525e1de9862f',
                                    'issue': 1065,
                                    'prerequisites': [],
                                    'depends': ['execution-dependencies']},
 'spell-longstrider': {'spell': 'Longstrider',
                       'page': 145,
                       'fixture': 'Action/touch/VSM/1h;+10ftSpeed,slotL '
                                  'targetsL;noWILLINGpredicateinvented,notConcentration. '
                                  'Negative:+20ft,permanence,expired,wrongtargetcount.',
                       'body_sha256': '61a2cd926caaf6381b96854031d0be256e87033c490060668eef9b2ecf3c1ee6',
                       'issue': 1066,
                       'prerequisites': [],
                       'depends': ['execution-dependencies']},
 'spell-meld-into-stone-entry': {'spell': 'Meld into Stone',
                                 'page': 148,
                                 'fixture': 'Action '
                                            'ORRitual/VS/touch/8h;stoneobject/surfaceMUSTfullyfitcasterbody;mergeequipmenttoo,nonmagicalsensescannotdetect. '
                                            'Trustedstonegeometryandprovenance;noinferredobjects.',
                                 'body_sha256': 'd7d062c90ddad7de6fd4730ef24ad267f0cd042451a3a792984037b484be53e7',
                                 'issue': 1067,
                                 'prerequisites': [],
                                 'depends': ['execution-dependencies']},
 'spell-meld-into-stone-inside': {'spell': 'Meld into Stone',
                                  'page': 148,
                                  'fixture': 'Acceptedmerge:nooutsidevision,hearPerceptionDisadvantage,awaretime,selfspellsonly,otherwisecannotmove;5ftmovementleaveatentryends. '
                                             'Negative:freeteleport,lookthroughstone,externalcasttargets.',
                                  'body_sha256': 'a1396934ec816222521a1b90f61eb303691d050918a88f3f1f919ebd4504b448',
                                  'issue': 1068,
                                  'prerequisites': [],
                                  'depends': ['execution-dependencies', 'spell-meld-into-stone-entry']},
 'spell-meld-into-stone-expulsion': {'spell': 'Meld into Stone',
                                     'page': 148,
                                     'fixture': 'MinorstonedamageNOcasterharm;partialdestruction/reshapetooSmall→expel6d6FORCE;completedestruction/transmutation→50FORCE. '
                                                'Closestunoccupiedspaceentry,Prone488. '
                                                'Trustedobjectmutationeventsonly,noguessedcollapse;rollback/reloadonce.',
                                     'body_sha256': 'f2a1c7eb86f9cd44e3b27c44fffc8ae46cdb9fcd1c2cf2af5985eff26fcb8c30',
                                     'issue': 1069,
                                     'prerequisites': [488],
                                     'depends': ['execution-dependencies', 'spell-meld-into-stone-entry']},
 'spell-message': {'spell': 'Message',
                   'page': 148,
                   'fixture': 'Action/S+M/120ft/1round;targetONLYhearswhisper/replyONLYcaster;throughsolidobject '
                              'ONLYfamiliar+knowbeyondbarrier. BlockmagicalSilence,1ftSTONE/METAL/WOOD,thinlead. '
                              'NotDetectMagic1inchmetalbarrier;noVcomponentinvented/noothershear.',
                   'body_sha256': '48ee0072ee96ceb2e2b50e85c57ba749de8c333c6e11a3cc1295650736f46660',
                   'issue': 1070,
                   'prerequisites': [],
                   'depends': ['execution-dependencies']},
 'spell-nondetection': {'spell': 'Nondetection',
                        'page': 151,
                        'fixture': 'Action/touch/VSM CONSUMEDdiamond25GP/8h;willingcreature '
                                   'ORplace/object<=10ftEACHdimension;cannotbetargetDivinationorperceivedmagicalscryingsensors. '
                                   'Notallmagicimmunity;negative:unwillingcreature,11ftobject,unpaidmaterial,nondetectionhidesordinarysight.',
                        'body_sha256': '882dcdd3ac4489bc8c6edc042e6401c3e3b14f12d6f4b0c0aa35c5e9b429b773',
                        'issue': 1071,
                        'prerequisites': [],
                        'depends': ['execution-dependencies']},
 'spell-pass-without-trace': {'spell': 'Pass without Trace',
                              'page': 151,
                              'fixture': 'Action/VSM/self/Concentration1h;30ftEmanation,caster+CHOSENcreatureswhileinside+10DEXSTEALTH '
                                         'andnotracks. '
                                         'NoautomaticHide/Invisible;outside/expiredlosescontribution;not+10allchecks.',
                              'body_sha256': 'a991fcc4bc258e23d973041d2738953f8c021cec0bceed785867ed5ac6bcaf95',
                              'issue': 1072,
                              'prerequisites': [],
                              'depends': ['execution-dependencies']},
 'spell-plant-growth-overgrowth': {'spell': 'Plant Growth',
                                   'page': 153,
                                   'fixture': 'Action/VS/150ft/instant;NORMALplants100ftradiusSphere,casterchosenexcludedareasanysize;movement4ft/ft,notordinary2ftdifficultterrain;noaddedplantswhereabsent.',
                                   'body_sha256': 'bff825e7cb3e6bfad935d7adf45675fefe0e508b26db7224a4ff6b124390cf10',
                                   'issue': 1073,
                                   'prerequisites': [],
                                   'depends': ['execution-dependencies']},
 'spell-plant-growth-enrichment': {'spell': 'Plant Growth',
                                   'page': 153,
                                   'fixture': '8hcasting/VS/point150ft;halfmileradiusplantsenriched365days,doublefoodharvest,ONLYonePlantGrowthperYEAR. '
                                              'Explicitauthoritativecrop/timefacts;noinventedfood/repeatmultiplier.',
                                   'body_sha256': '181628d54e0b68496c18f15bb3487a9cb100e79ee714449c0d0f41d00e2fc9ba',
                                   'issue': 1074,
                                   'prerequisites': [],
                                   'depends': ['execution-dependencies', 'spell-plant-growth-overgrowth']},
 'spell-poison-spray': {'spell': 'Poison Spray',
                        'page': 153,
                        'fixture': 'Action/VS/30ft;RANGEDSPELLATTACK '
                                   'creature,hit1d12Poison;TOTALlevels5/11/17→2/3/4d12. '
                                   'Negative:CONsaveoldvariant,missdamage,classlevelscaling,31ft.',
                        'body_sha256': '55d07b6d73f6542d1dc8f250a6bb81d75b6be5a15cb7b6287609c03c436ddbcc',
                        'issue': 1075,
                        'prerequisites': [],
                        'depends': ['execution-dependencies']},
 'spell-produce-flame-light': {'spell': 'Produce Flame',
                               'page': 156,
                               'fixture': 'BonusAction/VS/self/10min;handflameNOheat/ignition,Bright20+Dim20;RECASTendsold. '
                                          'Hurling doesnotconsume/endflame. PDF106ongoingtime-span dismissal '
                                          'allowedNOACTION onlyifnotIncapacitated; retain each ending trigger.',
                               'body_sha256': '87d1238e36afe37d5d48d2bf8fb86b6444fae595f872c368922d251b78b066d8',
                               'issue': 1076,
                               'prerequisites': [],
                               'depends': ['execution-dependencies']},
 'spell-produce-flame-attack': {'spell': 'Produce Flame',
                                'page': 156,
                                'fixture': 'Whileflameactive '
                                           'MagicACTIONrangedspellattackcreatureORobject60ft;hit1d8Fire,TOTAL5/11/17→2/3/4d8. '
                                           'Negative:BonusActionattack,61ft,spellendsafterattack,inventedignition.',
                                'body_sha256': '5ada30933fb1721553df5c30a34544a10d26ffcc54d93468bd5e3ad6e877e098',
                                'issue': 1077,
                                'prerequisites': [],
                                'depends': ['execution-dependencies', 'spell-produce-flame-light']},
 'spell-protection-from-energy': {'spell': 'Protection from Energy',
                                  'page': 157,
                                  'fixture': 'Action/touchwilling/VS/Concentration1h;ONEAcid/Cold/Fire/Lightning/ThunderResistance. '
                                             'Negative:Force/Poison/alltypes/unwilling,permanent.',
                                  'body_sha256': '3a741b40a56f8d553fe97e15b8e41f22e66a405d74a1a52abee36d25cfc31cba',
                                  'issue': 1078,
                                  'prerequisites': [],
                                  'depends': ['execution-dependencies']},
 'spell-shillelagh': {'spell': 'Shillelagh',
                      'page': 162,
                      'fixture': 'BonusAction/VSM/self/1min;HELDClub/Quarterstaff,MELEEattack+damage '
                                 'mayuseWISforRanger;dieD8,total5D10/11D12/17TWOd6;choiceFORCEorweaponnormaldamage. '
                                 'Recast/letgoends;notweapon+extraD8/unarmed/longsword.',
                      'body_sha256': 'c7a181811f41c6a9ebda59b3cec3f92d8b6410c5f0f750a6fce493abe1d3e99b',
                      'issue': 1079,
                      'prerequisites': [],
                      'depends': ['execution-dependencies']},
 'spell-silence': {'spell': 'Silence',
                   'page': 162,
                   'fixture': 'Action '
                              'ORRitual/VS/120ft/Concentration10min;20ftradiusSphere,nosoundwithin/through;ENTIRELYinsidecreature/objectThunderIMMUNE,creatureDeafened961,Verbalcastingimpossiblethere. '
                              'Negative:partialintersectionimmunity,allspellblock,noConcentration.',
                   'body_sha256': 'c05bc828635d539e329087ca1f0e8451ba900ebbaacf1e750aecaf7f83270d90',
                   'issue': 1080,
                   'prerequisites': [961],
                   'depends': ['execution-dependencies']},
 'spell-speak-with-plants-facts': {'spell': 'Speak with Plants',
                                   'page': 164,
                                   'fixture': 'Action/VS/self/10min;IMMOBILE30ftEmanation,limitedplantsentience/simplecommands,pastDAYauthorizedevents '
                                              'only;Plantcreaturescommunicateascommonlanguage. '
                                              'Noinventedknowledge/uproot/movement.',
                                   'body_sha256': '57d17a628ccffb2287ca9545699e042a4b5ff81deb2fe9a91a2dfb6f2f576480',
                                   'issue': 1081,
                                   'prerequisites': [],
                                   'depends': ['execution-dependencies']},
 'spell-speak-with-plants-terrain': {'spell': 'Speak with Plants',
                                     'page': 164,
                                     'fixture': 'Whileacceptedspellactive changeplantDifficultTerrain→ordinary '
                                                'ORordinaryplantpresentterrain→DifficultforDURATION;branches/tendrils/stalksmaymove,notuproot. '
                                                'Noabsentplants/lastingterraform;restoreonlysourcecontribution.',
                                     'body_sha256': '6f0e61d1e06b3303ae9334c975c9def8a8526f74d4edf1f012c96f5e8e9d144c',
                                     'issue': 1082,
                                     'prerequisites': [],
                                     'depends': ['execution-dependencies', 'spell-speak-with-plants-facts']},
 'spell-spike-growth-terrain': {'spell': 'Spike Growth',
                                'page': 164,
                                'fixture': 'Action/VSM/150ft/Concentration10min;20ftradiusSpheregroundDifficultTerrain,camouflagednatural;creaturecouldnotseecast '
                                           'mustSearchActionWISPerception ORSurvival vsDC recognizebeforeentry. '
                                           'Notautomaticlocationdisclosure.',
                                'body_sha256': 'b28fd654d38c1f3fb73da98bb48a72b80f42974f6674cc18e92846263e71d8aa',
                                'issue': 1083,
                                'prerequisites': [],
                                'depends': ['execution-dependencies']},
 'spell-spike-growth-movement': {'spell': 'Spike Growth',
                                 'page': 164,
                                 'fixture': 'Sourcearea movementINTOorWITHIN EACH5ft→2d4Piercing,notonceperturn. '
                                            'Saveactualdistanceanddice,retryonce,nostandingdamage/enteringonly/30ftsingletick. '
                                            'Consumesameauthoritativeterraingeometry.',
                                 'body_sha256': '2c89e5ed7f784ad4e2bf07e33d11690231cc3810edf6e37d12e1e53c7d842ef5',
                                 'issue': 1084,
                                 'prerequisites': [],
                                 'depends': ['execution-dependencies', 'spell-spike-growth-terrain']},
 'spell-starry-wisp': {'spell': 'Starry Wisp',
                       'page': 165,
                       'fixture': 'Action/VS/60ft;creatureORobjectrangedspellattackhit1d8Radiant,total5/11/17→2/3/4d8;untilENDYOURNEXTturnDim10ft '
                                  'andCANNOTBENEFITInvisible962 (suppression,notdeleteallsources). '
                                  'Negative:missrider,classlevelscaling,permanentending.',
                       'body_sha256': 'a4d49a449072c022c8e1203c05f3a115dfbb3184ecacf814cce999792431b152',
                       'issue': 1085,
                       'prerequisites': [962],
                       'depends': ['execution-dependencies']},
 'spell-stoneskin': {'spell': 'Stoneskin',
                     'page': 165,
                     'fixture': 'Action/touchwilling/VSM '
                                'consumeddiamond100GP/Concentration1h;ResistanceBludgeoning/Piercing/Slashing '
                                'INCLUDINGmagic. Negative:nonmagiconlyoldvariant,unpaid,unwilling,alltypes.',
                     'body_sha256': '27b48bd7e866e62500570a9e68482799ef312ca379fbe99f058ce5a0da9d7948',
                     'issue': 1086,
                     'prerequisites': [],
                     'depends': ['execution-dependencies']},
 'spell-water-breathing': {'spell': 'Water Breathing',
                           'page': 173,
                           'fixture': 'Action '
                                      'ORRitual/VSM/30ft/24h;uptoTENwillingcreaturesunderwaterbreathingretainNORMALrespiration. '
                                      'Negative:11/unwilling/31ft/losingairbreathing/inferredwaterfacts.',
                           'body_sha256': 'd114095031adaad936cb745a1be6c6805cf93d2b4e18095938d1e83937f5193c',
                           'issue': 1087,
                           'prerequisites': [],
                           'depends': ['execution-dependencies']},
 'spell-water-walk': {'spell': 'Water Walk',
                      'page': 173,
                      'fixture': 'Action ORRitual/VSM/30ft/1h;uptoTENwillingcreaturescrossliquids '
                                 'asharmlesssolidground,LAVAHEATstilldamages;BonusActionpasssurface→liquidORback,fallpassesthrough. '
                                 'Noautoemerge/freeBonusAction/heatimmunity.',
                      'body_sha256': 'f1787239a74773d61aa2762cf8f02777a46777b7e572099a9788519706d73e73',
                      'issue': 1088,
                      'prerequisites': [],
                      'depends': ['execution-dependencies']},
 'spell-wind-wall-placement': {'spell': 'Wind Wall',
                               'page': 174,
                               'fixture': 'Action/VSM/120ft/Concentration1min;continuousgroundpath<=50long15high1thick;trustedgeometry. '
                                          'Negative:floatingwall,gaps,51long,overlapinferred.',
                               'body_sha256': '77802652f47b8b028f5af5b7397aee668416c0b1d8655394cacf083ea661e28e',
                               'issue': 1089,
                               'prerequisites': [],
                               'depends': ['execution-dependencies']},
 'spell-wind-wall-damage': {'spell': 'Wind Wall',
                            'page': 174,
                            'fixture': 'APPEARANCEonlyeachcreatureinareaSTRsave4d8Bludgeoningfail/halfsuccess. '
                                       'Negative:entrytick,DEXsave,secondappearance,retrydamage.',
                            'body_sha256': '7a3b7e64e90e0aedadf5240afd06eb2001f71f8e82a717a7d8bad1c853bf9e40',
                            'issue': 1090,
                            'prerequisites': [],
                            'depends': ['execution-dependencies', 'spell-wind-wall-placement']},
 'spell-wind-wall-barrier': {'spell': 'Wind Wall',
                             'page': 174,
                             'fixture': 'Activewallkeepsfog/smoke/gasesatbay;SmallorfewerFLYINGcreatures/objectscannotpass,gaseouscreaturescannotpass;looseLIGHTmaterialsflyup. '
                                        'Ordinaryprojectilesbehindwallautomiss;Giant/siegeBOULDERS unaffected. '
                                        'Notallattacksblocked,nogroundedSmallban.',
                             'body_sha256': 'a2f9d9a852c734a5ddc184cbb079cef0f94b9a15e418aec0f10a99a45312587d',
                             'issue': 1091,
                             'prerequisites': [],
                             'depends': ['execution-dependencies', 'spell-wind-wall-placement']}}

# Record context is part of the identity: repeated class-level keys have distinct owners.
SOURCE_OWNERS = {'class/ranger': {'primary-ability/dexterity': 'core',
                  'primary-ability/wisdom': 'core',
                  'hit-die-sides': 'hp',
                  'srd-5.2.1:mechanic/saving-throws/strength': 'core',
                  'srd-5.2.1:mechanic/saving-throws/dexterity': 'core',
                  'srd-5.2.1:mechanic/weapons/simple': 'core',
                  'srd-5.2.1:mechanic/weapons/martial': 'core',
                  'srd-5.2.1:mechanic/armor/light': 'core',
                  'srd-5.2.1:mechanic/armor/medium': 'core',
                  'srd-5.2.1:mechanic/armor/shields': 'core',
                  'level-one/all-core-traits-and-features': 'level-1',
                  'multiclass/hit-die-martial-weapons-one-class-skill-light-medium-shields': 'later-entry',
                  'multiclass/level-one-features-multiclass-slot-rules': 'later-entry'},
 'choice/ranger/skills': {'skills': 'skills',
                          'srd-5.2.1:mechanic/skills/animal-handling': 'skills',
                          'srd-5.2.1:mechanic/skills/athletics': 'skills',
                          'srd-5.2.1:mechanic/skills/insight': 'skills',
                          'srd-5.2.1:mechanic/skills/investigation': 'skills',
                          'srd-5.2.1:mechanic/skills/nature': 'skills',
                          'srd-5.2.1:mechanic/skills/perception': 'skills',
                          'srd-5.2.1:mechanic/skills/stealth': 'skills',
                          'srd-5.2.1:mechanic/skills/survival': 'skills'},
 'choice/ranger/multiclass-skill': {'multiclass-skill': 'later-entry',
                                    'srd-5.2.1:mechanic/skills/animal-handling': 'later-entry',
                                    'srd-5.2.1:mechanic/skills/athletics': 'later-entry',
                                    'srd-5.2.1:mechanic/skills/insight': 'later-entry',
                                    'srd-5.2.1:mechanic/skills/investigation': 'later-entry',
                                    'srd-5.2.1:mechanic/skills/nature': 'later-entry',
                                    'srd-5.2.1:mechanic/skills/perception': 'later-entry',
                                    'srd-5.2.1:mechanic/skills/stealth': 'later-entry',
                                    'srd-5.2.1:mechanic/skills/survival': 'later-entry'},
 'prerequisite/ranger/multiclass-entry': {'multiclass-entry-only': 'later-entry'},
 'choice/ranger/starting-equipment': {'starting-equipment': 'equipment',
                                      'srd-5.2.1:mechanic/equipment/studded-leather-armor': 'equipment',
                                      'srd-5.2.1:mechanic/equipment/scimitar': 'equipment',
                                      'srd-5.2.1:mechanic/equipment/shortsword': 'equipment',
                                      'srd-5.2.1:mechanic/equipment/longbow': 'equipment',
                                      'srd-5.2.1:mechanic/equipment/arrows': 'equipment',
                                      'srd-5.2.1:mechanic/equipment/quiver': 'equipment',
                                      'srd-5.2.1:mechanic/equipment/druidic-focus-sprig-of-mistletoe': 'equipment',
                                      'srd-5.2.1:mechanic/equipment/explorers-pack': 'equipment',
                                      'srd-5.2.1:mechanic/currency/gp': 'equipment',
                                      'equipment/a': 'equipment',
                                      'equipment/b': 'equipment'},
 'level/ranger/1': {'class-progression': 'level-1'},
 'feature/ranger/spellcasting': {'spellcasting/wisdom': 'spellcasting',
                                 'spellcasting/druidic-focus': 'spellcasting',
                                 'spellcasting/always-prepared-excluded-still-ranger-spells': 'spellcasting',
                                 'recommended-spell/cure-wounds': 'spellcasting',
                                 'recommended-spell/ensnaring-strike': 'spellcasting',
                                 'spellcasting/long-rest-replace-one-eligible-spell': 'spellcasting',
                                 'prepared-replacements': 'spellcasting',
                                 'feature/spellcasting': 'spellcasting'},
 'feature/ranger/favored-enemy': {'favored-enemy/source-5-2-1-always-prepared-hunters-mark': 'favored-enemy',
                                  'srd-5.2.1:mechanic/spells/hunters-mark': 'favored-enemy',
                                  'favored-enemy/no-slot-cast': 'favored-enemy',
                                  'uses-per-cast': 'favored-enemy',
                                  'feature/favored-enemy': 'favored-enemy'},
 'feature/ranger/weapon-mastery': {'mastery/long-rest-replace-kinds': 'weapon-mastery',
                                   'feature/weapon-mastery': 'weapon-mastery'},
 'level/ranger/2': {'class-progression': 'level-2'},
 'feature/ranger/deft-explorer': {'feature/deft-explorer': 'deft-explorer'},
 'feature/ranger/fighting-style': {'druidic-warrior/replace-one-on-ranger-level': 'druidic-cantrips',
                                   'cantrip-replacements': 'druidic-cantrips',
                                   'feature/fighting-style': 'fighting-style'},
 'level/ranger/3': {'class-progression': 'level-3'},
 'feature/ranger/ranger-subclass': {'subclass/features-at-or-below-ranger-level': 'ranger-subclass',
                                    'feature/ranger-subclass': 'ranger-subclass'},
 'level/ranger/4': {'class-progression': 'level-4'},
 'feature/ranger/ability-score-improvement': {'asi-repeat-at-8-12-16': 'ability-score-improvement',
                                              'feature/ability-score-improvement': 'ability-score-improvement'},
 'level/ranger/5': {'class-progression': 'level-5'},
 'feature/ranger/extra-attack': {'extra-attack/own-turn-attack-action': 'extra-attack',
                                 'attacks': 'extra-attack',
                                 'feature/extra-attack': 'extra-attack'},
 'level/ranger/6': {'class-progression': 'level-6'},
 'feature/ranger/roving': {'roving/speed-increase-when-not-heavy-armor': 'roving',
                           'speed-bonus-feet': 'roving',
                           'roving/climb-and-swim-equal-speed': 'roving',
                           'feature/roving': 'roving'},
 'level/ranger/7': {'class-progression': 'level-7', 'subclass-feature': 'level-7'},
 'level/ranger/8': {'class-progression': 'level-8'},
 'level/ranger/9': {'class-progression': 'level-9'},
 'feature/ranger/expertise': {'feature/expertise': 'expertise'},
 'level/ranger/10': {'class-progression': 'level-10'},
 'feature/ranger/tireless': {'tireless/magic-action-self-temporary-hp': 'tireless-hp',
                             'tireless/temp-hp-die-plus-wisdom-minimum-one': 'tireless-hp',
                             'temporary-hp-dice': 'tireless-hp',
                             'temporary-hp-die-sides': 'tireless-hp',
                             'temporary-hp-minimum': 'tireless-hp',
                             'tireless/short-rest-decrease-exhaustion-if-any': 'tireless-rest',
                             'exhaustion-reduction': 'tireless-rest',
                             'feature/tireless': 'tireless-hp'},
 'level/ranger/11': {'class-progression': 'level-11', 'subclass-feature': 'level-11'},
 'level/ranger/12': {'class-progression': 'level-12'},
 'level/ranger/13': {'class-progression': 'level-13'},
 'feature/ranger/relentless-hunter': {'relentless-hunter/damage-cannot-break-hunters-mark-concentration': 'relentless-hunter',
                                      'feature/relentless-hunter': 'relentless-hunter'},
 'level/ranger/14': {'class-progression': 'level-14'},
 'feature/ranger/natures-veil': {'veil/bonus-action-self-invisible-until-end-next-turn': 'natures-veil',
                                 'feature/natures-veil': 'natures-veil'},
 'level/ranger/15': {'class-progression': 'level-15', 'subclass-feature': 'level-15'},
 'level/ranger/16': {'class-progression': 'level-16'},
 'level/ranger/17': {'class-progression': 'level-17'},
 'feature/ranger/precise-hunter': {'precise-hunter/advantage-on-attacks-against-own-current-mark': 'precise-hunter',
                                   'feature/precise-hunter': 'precise-hunter'},
 'level/ranger/18': {'class-progression': 'level-18'},
 'feature/ranger/feral-senses': {'feral-senses/blindsight': 'feral-senses',
                                 'blindsight-feet': 'feral-senses',
                                 'feature/feral-senses': 'feral-senses'},
 'level/ranger/19': {'class-progression': 'level-19'},
 'feature/ranger/epic-boon': {'recommended-feat/boon-of-dimensional-travel': 'epic-boon',
                              'feature/epic-boon': 'epic-boon'},
 'level/ranger/20': {'class-progression': 'level-20'},
 'feature/ranger/foe-slayer': {'foe-slayer/replace-hunters-mark-damage-die': 'foe-slayer',
                               'original-die-sides': 'foe-slayer',
                               'replacement-die-sides': 'foe-slayer',
                               'feature/foe-slayer': 'foe-slayer'},
 'choice/ranger/prepared-spells': {'spellcasting/prepare-to-table-count-on-increase': 'spellcasting',
                                   'srd-5.2.1:mechanic/spells/ranger-list': 'spellcasting'},
 'prerequisite/ranger/prepared-slot-level': {'spell-level-with-available-slot': 'spellcasting'},
 'resource/ranger/spell-slot-1': {'spell-slot-1/long-rest-recover-all': 'slots'},
 'resource/ranger/spell-slot-2': {'spell-slot-2/long-rest-recover-all': 'slots'},
 'resource/ranger/spell-slot-3': {'spell-slot-3/long-rest-recover-all': 'slots'},
 'resource/ranger/spell-slot-4': {'spell-slot-4/long-rest-recover-all': 'slots'},
 'resource/ranger/spell-slot-5': {'spell-slot-5/long-rest-recover-all': 'slots'},
 'resource/ranger/favored-enemy': {'favored-enemy/long-rest-recover-all': 'favored-enemy'},
 'choice/ranger/weapon-mastery': {'weapon-mastery': 'weapon-mastery',
                                  'srd-5.2.1:mechanic/weapons/proficient-kinds': 'weapon-mastery'},
 'prerequisite/ranger/mastery-proficiency': {'selected-weapon-proficient': 'weapon-mastery'},
 'choice/ranger/deft-explorer-expertise': {'deft-explorer-expertise': 'deft-explorer',
                                           'srd-5.2.1:mechanic/skills/proficient-without-expertise': 'deft-explorer'},
 'prerequisite/ranger/deft-explorer-eligibility': {'proficient-skill-lacking-expertise': 'deft-explorer'},
 'choice/ranger/expertise-expertise': {'expertise-expertise': 'expertise',
                                       'srd-5.2.1:mechanic/skills/proficient-without-expertise': 'expertise'},
 'prerequisite/ranger/expertise-eligibility': {'proficient-skill-lacking-expertise': 'expertise'},
 'choice/ranger/languages': {'languages': 'deft-explorer',
                             'srd-5.2.1:mechanic/languages/character-creation-tables': 'deft-explorer'},
 'choice/ranger/fighting-style': {'fighting-style': 'fighting-style',
                                  'srd-5.2.1:mechanic/feats/fighting-style': 'fighting-style',
                                  'druidic-warrior': 'druidic-cantrips'},
 'choice/ranger/druidic-warrior-cantrips': {'druidic-warrior-cantrips': 'druidic-cantrips',
                                            'druidic-warrior/count-as-ranger-spells-wisdom': 'druidic-cantrips',
                                            'recommended-cantrip/guidance': 'druidic-cantrips',
                                            'recommended-cantrip/starry-wisp': 'druidic-cantrips',
                                            'srd-5.2.1:mechanic/spells/druid-cantrips': 'druidic-cantrips'},
 'prerequisite/ranger/druidic-warrior-selected': {'druidic-warrior-selected': 'druidic-cantrips'},
 'choice/ranger/subclass': {'subclass': 'ranger-subclass',
                            'srd-5.2.1:mechanic/subclasses/ranger': 'ranger-subclass',
                            'srd-5.2.1:mechanic/subclasses/ranger/hunter': 'ranger-subclass'},
 'choice/ranger/ability-score-improvement': {'ability-score-improvement': 'ability-score-improvement',
                                             'srd-5.2.1:mechanic/feats/ability-score-improvement': 'ability-score-improvement',
                                             'srd-5.2.1:mechanic/feats/qualified': 'ability-score-improvement'},
 'prerequisite/ranger/ability-score-improvement-qualified': {'selected-feat-qualified': 'ability-score-improvement'},
 'choice/ranger/epic-boon': {'epic-boon': 'epic-boon',
                             'srd-5.2.1:mechanic/feats/epic-boon': 'epic-boon',
                             'srd-5.2.1:mechanic/feats/qualified': 'epic-boon'},
 'prerequisite/ranger/epic-boon-qualified': {'selected-feat-qualified': 'epic-boon'},
 'prerequisite/ranger/roving-armor': {'not-wearing-heavy-armor': 'roving'},
 'resource/ranger/tireless': {'tireless/long-rest-recover-all': 'tireless-hp',
                              'wisdom-modifier-minimum-one': 'tireless-hp'},
 'prerequisite/ranger/tireless-exhaustion': {'exhaustion-present': 'tireless-rest'},
 'prerequisite/ranger/relentless-mark': {'concentrating-on-hunters-mark': 'relentless-hunter'},
 'resource/ranger/natures-veil': {'natures-veil/long-rest-recover-all': 'natures-veil',
                                  'wisdom-modifier-minimum-one': 'natures-veil'},
 'prerequisite/ranger/precise-mark': {'target-is-own-current-hunters-mark': 'precise-hunter'},
 'prerequisite/ranger/foe-slayer-mark': {'hunters-mark-damage': 'foe-slayer'},
 'subclass/ranger/hunter/identity': {'hunter/subclass': 'hunter-identity'},
 'prerequisite/ranger/hunter/identity-level': {'hunter/class-level': 'hunter-identity'},
 'feature/ranger/hunter/hunters-lore': {'hunter/hunters-lore/know-whether-marked-creature-has-immunities-resistances-vulnerabilities-and-what-they-are': 'hunters-lore'},
 'prerequisite/ranger/hunter/hunters-lore-level': {'hunter/class-level': 'hunters-lore'},
 'prerequisite/ranger/hunter/hunters-lore-mark': {'hunter/target-marked-by-your-hunters-mark': 'hunters-lore'},
 'feature/ranger/hunter/hunters-prey': {'hunter/hunters-prey/gain-one-chosen-option': 'hunters-prey'},
 'prerequisite/ranger/hunter/hunters-prey-level': {'hunter/class-level': 'hunters-prey'},
 'choice/ranger/hunter/hunters-prey': {'hunter/choose-one': 'hunters-prey',
                                       'hunter/replace-chosen-option-with-other-on-finish-short-or-long-rest': 'hunters-prey',
                                       'hunter/colossus-slayer': 'colossus-slayer',
                                       'hunter/horde-breaker': 'horde-breaker'},
 'prerequisite/ranger/hunter/colossus-slayer': {'hunter/colossus-slayer/only-while-chosen': 'colossus-slayer',
                                                'hunter/colossus-slayer/on-weapon-hit-creature': 'colossus-slayer',
                                                'hunter/colossus-slayer/weapon-extra-damage-to-hit-target': 'colossus-slayer',
                                                'hunter/colossus-slayer/only-if-target-missing-any-hit-points': 'colossus-slayer',
                                                'hunter/colossus-slayer/once-per-turn': 'colossus-slayer',
                                                'hunter/colossus-slayer/damage-dice': 'colossus-slayer',
                                                'hunter/colossus-slayer/damage-die-sides': 'colossus-slayer'},
 'prerequisite/ranger/hunter/horde-breaker': {'hunter/horde-breaker/only-while-chosen': 'horde-breaker',
                                              'hunter/horde-breaker/once-on-each-of-your-turns-when-making-weapon-attack': 'horde-breaker',
                                              'hunter/horde-breaker/optional-another-attack-with-same-weapon': 'horde-breaker',
                                              'hunter/horde-breaker/target-different-creature-within-weapon-range-not-attacked-this-turn': 'horde-breaker',
                                              'hunter/horde-breaker/target-within-feet-of-original-target': 'horde-breaker'},
 'feature/ranger/hunter/defensive-tactics': {'hunter/defensive-tactics/gain-one-chosen-option': 'defensive-tactics'},
 'prerequisite/ranger/hunter/defensive-tactics-level': {'hunter/class-level': 'defensive-tactics'},
 'choice/ranger/hunter/defensive-tactics': {'hunter/choose-one': 'defensive-tactics',
                                            'hunter/replace-chosen-option-with-other-on-finish-short-or-long-rest': 'defensive-tactics',
                                            'hunter/escape-the-horde': 'escape-the-horde',
                                            'hunter/multiattack-defense': 'multiattack-defense'},
 'prerequisite/ranger/hunter/escape-the-horde': {'hunter/escape-the-horde/only-while-chosen': 'escape-the-horde',
                                                 'hunter/escape-the-horde/opportunity-attacks-against-you-have-disadvantage': 'escape-the-horde'},
 'prerequisite/ranger/hunter/multiattack-defense': {'hunter/multiattack-defense/only-while-chosen': 'multiattack-defense',
                                                    'hunter/multiattack-defense/when-creature-hits-you-with-attack-roll': 'multiattack-defense',
                                                    'hunter/multiattack-defense/that-creature-disadvantage-on-all-other-attack-rolls-against-you-this-turn': 'multiattack-defense'},
 'feature/ranger/hunter/superior-hunters-prey': {'hunter/superior-hunters-prey/once-per-turn-when-you-deal-damage-to-marked-creature': 'superior-hunters-prey',
                                                 'hunter/superior-hunters-prey/optional-deal-that-spells-extra-damage': 'superior-hunters-prey',
                                                 'hunter/superior-hunters-prey/target-different-creature-you-can-see': 'superior-hunters-prey',
                                                 'hunter/superior-hunters-prey/target-within-feet-of-first-creature': 'superior-hunters-prey'},
 'prerequisite/ranger/hunter/superior-hunters-prey-level': {'hunter/class-level': 'superior-hunters-prey'},
 'prerequisite/ranger/hunter/superior-hunters-prey-mark': {'hunter/target-marked-by-your-hunters-mark': 'superior-hunters-prey'},
 'prerequisite/ranger/hunter/superior-hunters-prey-damage': {'hunter/superior-hunters-prey/use-current-canonical-hunters-mark-extra-damage': 'superior-hunters-prey'},
 'feature/ranger/hunter/superior-hunters-defense': {'hunter/superior-hunters-defense/when-you-take-damage': 'superior-hunters-defense',
                                                    'hunter/superior-hunters-defense/optional-reaction': 'superior-hunters-defense',
                                                    'hunter/superior-hunters-defense/self-resistance-to-triggering-damage-and-any-other-damage-of-same-type': 'superior-hunters-defense',
                                                    'hunter/superior-hunters-defense/until-end-of-current-turn': 'superior-hunters-defense'},
 'prerequisite/ranger/hunter/superior-hunters-defense-level': {'hunter/class-level': 'superior-hunters-defense'}}

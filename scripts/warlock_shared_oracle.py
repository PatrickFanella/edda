"""Independent bounded effect predicates for the Warlock-only reservation.

Mind Spike ownership is explicitly Warlock; disputed Sorcerer admission is absent.
Existing named spell owners and Sorcerer provisional owners remain external.
"""

# Each part is one future implementation PR. These are effect scopes, never
# permissions to cast or duplicate at-will/Arcanum/Tome variants.
SPELLS = {
 'Astral Projection': (111, {
  'projection': 'Level9;1hour;10ft; consumes1000GPjacinth+100GPsilver PERtarget; self+up to8willingcreatures. Bodies Unconscious/nofood/air/aging; astral copies stats/possessions. Positive9total/9900GP; reject10targets, unwilling, missingper-targetcost; alreadyAstralPlane endsinstantly.',
  'cord-travel': 'Cord visible1ft, can be cut ONLY by explicit sourceeffect; cut kills body+astralform. LeavingAstralPlane pulls body/possessions alongcord to reenter. Positive explicitcordcut killsboth; reject ordinaryweaponcut, arbitraryplanetravel without realworldowner or privatecoordinates.',
  'damage-ending': 'Body/astralform effects separate; either0HP ends only thattarget, nondead returntoownbody and end suspendedanimation. CasterMagicAction dismissesall; preserve general106noactiondismissal applicable exception reconciliation. Positiveone0HP return/othersremain; reject sharedHP, killing livingbody on ordinaryend, alltargetsending fromone0HP.',
 }),
 'Bane': (112, {'save-penalty': 'Level1Action30ft,VSMblood,Concentration1minute; up to3visiblecreatures CHA save; failure subtract1d4 each attackroll/save, +1target perslotabove1. Positive slot3targets5,d4=3 reducesattack17->14; reject6targets, abilitycheckpenalty, successfulsavepenalty, stacking same-spell penalties.'}),
 'Befuddlement': (112, {
  'damage-block': 'Level8Action150ft,VSMkeyring,instant;1visiblecreature INTsave; failure10d12Psychic and cannotcastspells/takeMagicAction; successhalfdamageonly. Positive10rolls12=120/60success; reject successblock,9d12, restrictionallactions or immunityinferredfromIntelligence.',
  'ending': 'At end every30days repeatINTsave, successendsblock; GreaterRestoration/Heal/Wish exactoutputs endblock. Positive day30DC18/roll19 restorescasting; actual621+623 GreaterRestoration on same known recipient withExhaustion1 removes thatExhaustion AND ownBefuddlement source exception; actual585Wish binds exactselectedrecipient. Preserve old621five-option dispatcher; no blanketcure. Heal producer remainsUNBOUND/unavailable untilactualnamedqualifiedoutput. Rejectroll17,dailyrepeat,restcure,foreignrecipient,DispelMagic ofinstantaneousspell inferred or clientrestorationname.',
 }),
 'Contact Other Plane': (119, {
  'contact-questions': 'Level5;1minuteorRitual;Self,V;duration1minute; DC15INTsave success asksup to5questions beforeend; knowledgeableotherplanarentity, one-wordanswers or shortphrase ifmisleading. Positive5questions; reject6, latequestion, inventedknowledge or undisclosedworldfacts guessed as answer. OwnContactPatron exception separate.',
  'failed-contact': 'FailedDC15INTsave deals6d6Psychic and Incapacitated untilTHISCASTERLongRest via487/428; GreaterRestoration exactoutputends THISsource-boundIncappacket only. Positive6rolls6=36; same-recipient actual621+623 cureevent withExhaustion1 endsownpacketwithoutchangingoldfive-optiondispatcher. RejectShortRestend,otheractorrest,failedsavequestions,wrongdamage,foreigncureevent or treatingallnormalcastsauto-success.',
 }),
 'Create Undead': (121, {
  'animate': 'Level6;1minute;10ft,VSM150GPonyx percorpse(nonconsumed); ONLYnight; up to3Small/MediumHumanoidcorpses becomeGhouls. Positive3corpses/450GPstones; reject daytime, Largecorpse, nonHumanoid, consumedstones or metadata-onlyGhoul.',
  'command': 'BonusAction eachownturn, mentallycommand anyownedanimatedcreatures within120ft, samemultiplecommand; choose nextturnaction/move or generalcommand; taskcontinuesuntilcomplete. Uncommanded Dodgeandonlyavoidharmawaymovement via782. Positive120ft command; reject121ft, freeBA or unqualifiedattack.',
  'control-renewal': 'Control24hours then stopsobeying; recast BEFOREexpiry reassertscontrol ratherthancreatesnew. Positive23h59mrenew24h; reject afterexpiryrenewasowned, newcreaturecreation inrenewmode or destructionat24hinferred.',
  'upcast': 'Slot7 fourGhouls;8 fiveGhouls ORtwoGhasts/Wights;9 sixGhouls ORthreeGhasts/Wights ORtwoMummies. Positive9twoMummies; reject slot8Mummies, mixedunboundedroster or unavailable namedstatblocks. Eachform requires actualnamedboundedsourceowner.',
 }),
 'Dream': (126, {
  'messenger': 'Level5;1minute;Special;VSMsand;8hours;knowncreature sameplane, selforwillingtouchedmessengertrance Incapacitated/Speed0. Ifasleep converseandshapedream, perfectrecallwaking; messengerendsatwill. Positivewillingtouch; reject crossplane, unwillingmessenger, using illusorydreamobjects inrealworld.',
  'awake-wait': 'Awaketarget: messengerknowsawake and mayendtrance/spell or waitforsleep. Positivewake->sleep entersdreamwithinduration; reject forcedsleep, extending8hbywaiting or leakingotherprivatefacts.',
  'nightmare': 'Terrifyingmessenger delivers atmost10words thenWISsave; failure gainsnorestbenefit and takes3d6Psychicwhenwakes. Positive10words/3rolls6=18; reject11words, successdamage, damagebeforewaking, denyingrest on normalspeech.',
 }),
 'Eldritch Blast': (127, {'beams': 'Action120ft,VS,instant; rangedspellattack againstcreature ORobject, hit1d10Force; characterlevels5/11/17 give2/3/4beams, eachseparateattack, sameordifferenttargets. Positivelevel17fourrolls/fourhits40with10s; reject fifthbeam, oneattackrollforall, Warlocklevelscaling or savingthrow. Invocationdamage/range/push modifiers separate.'}),
 'Enthrall': (128, {'perception': 'Level2Action60ft,VS,Concentration1minute; anychosenvisiblecreaturesWISsave; creaturesyou/companionsfighting autosucceed; failedsave-10WISPerceptionchecks ANDPassivePerception. PositivePassive15->5; reject combatantfailure, allWISchecks, hearing-onlytarget or Charmedcondition inferred.'}),
 'Floating Disk': (133, {
  'manifest-load': 'Level1ActionorRitual30ft,VSMmercury,1hour; seenunoccupiedspace forceplane3ftdiameter/1inchthick/3ftaboveground, capacity500lb. Above500endandcontentsfall. Positive500holds/501ends; rejectoccupiedspace, floatinghigher or forcewallinferred.',
  'follow': 'Immobilecasterwithin20ft; beyond20followsremainwithin20, crossesuneventerrain/stairs/slopes but NOelevationchange10ft+. Positive9ftstepfollows; reject10ftpitcrossing, manualride-drivenmovement or independentmovementwhenwithin20.',
  'distance-ending': 'Castermorethan100ftendsdisk. Positive101ftendandrealcontentsfall; reject100ftend, obstacleteleport, orphanedload or generaldismissalbanbyomission; qualifiedworldload/gravityoutputsrequired.',
 }),
 'Forcecage': (134, {
  'prison': 'Level7Action100ft,VSM1500GPrubydustconsumed,Concentration1hour; immobileInvisiblecube: cageupto20ftbars0.5inchdiameter/spacing ORsolidboxupto10ftblocksallmatter/spellsin/out. Positive20ftcage/10ftbox; reject11ftbox, movingcage, no1500GPcost.',
  'occupancy': 'Fullyinsidecreaturetrapped; partial/tooLargepushedfromcenteruntilfullyoutside. Positivepartialcreaturepushedoutside; reject trappingpartial, guessingunknownoccupancy or deletingoccupant. Actualboundedgeometry/worldownerneeded.',
  'escape': 'Nomundaneescape; teleport/interplanartravelattempt firstCHAsave, successpermitsmagicexit, failurewastesspell/effect; extendEtherealblocking. Positivefailedsavewastescost/noexit; reject unqualifiedteleport, mundanewalkout, saveafterexit or freefailure.',
  'ending': 'Concentration/expiry/casterdismissal viaqualifiedprotocol; specifically DispelMagiccannotdispelForcecage. Positiveconcentrationendremovesprison; reject level9DispelMagicending or copyinggenericobjectdamage destruction.',
 }),
 'Foresight': (134, {'benefits-ending': 'Level9;1minute;Touchwilling,VSMfeather;8hours; AdvantageallD20Tests, othercreaturesDisadvantageattackrollsagainsttarget; endsoldspellwhenyoucastagain. Positivecheck/save/attackAdvantage; reject unwilling, damageAdvantage, oldtargetretainedafterrecast or no1minutecastingconcentration.'}),
 'Glibness': (136, {
  'check-floor': 'Level8ActionSelf,V,1hour; whenownCharismacheck optionallyreplacerollednumber15. PositiveCHAcheckd20=4->15beforemodifiers; rejectDEXcheck, CHAsave, replacingtotal15 or forcedreplace20.',
  'truth-magic': 'For durationmagicdeterminingtruth indicatesselftruthfulregardlessspokencontent; retainhonestsourceprivateauthority. Positiverealtruth-detectionprotocolreports truthful; reject changinglistenersbeliefs, mundanedeceptionautowin or fabricatingunknowntruthspellbehavior.',
 }),
 'Hallucinatory Terrain': (139, {
  'terrain': 'Level4;10minutes;300ft;VSMmushroom;24hours;150ftCube naturalterrain looks/sounds/smellsanothernaturalterrain, tactileunchanged; manufacturedstructures/equipment/creaturesunchanged. Positivefieldlookspond; rejectrealwater, changedbuilding/creature or151ftCube.',
  'discern': 'Obvioustouchcanrevealillusion; otherwiseStudyINTInvestigationvsownspellDC disbelieves; discernedseesvagueimageoverrealterrain. PositiveDC15roll15reveals; reject freeall-observerdisclosure, realtactilealteration or perceptionchecksubstitution.',
 }),
 'Hellish Rebuke': (140, {'reaction-damage': 'Level1Reactiontakingdamagefromvisiblecreaturewithin60ft;VS,instant; damagingsourceDEXsave2d10Fire/halfsuccess,+1d10perslotabove1. Positive slot3four10s=40/20; reject61ft, unseen, misswithoutdamage, wrongdamagingsource, spentReaction or Firetargetimmunityascastingban.'}),
 'Hex': (140, {
  'curse-hit': 'Level1BonusAction90ft,visiblecreature,VSMnewteye,Concentration1hour; hitwithattackrolladds1d6Necrotic;choose1ability checksDisadvantage. Positive2attackhits2d6; reject savetrigger, chosenabilitySAVEdisadvantage or extraothercreaturedamage.',
  'retarget': 'Target0HPbeforeend permitsBonusActiononLATERturn curseanothercreature. Positive nextturnBAretarget; reject sameturn, targetalive, noBA or resettingduration. Same-spellnonstackingrequired.',
  'duration': 'Slot2upto4hours,3–4upto8hours,5+24hoursConcentration. PositivePact5duration24h; rejectslot3=24h, slot2=8h or noConcentration. Higher slotdoesnotincreased6damage.',
 }),
 'Hideous Laughter': (140, {
  'admission': 'Level1Action30ft,visiblecreature,VSMtart/feather,Concentration1minute; WISfailure Prone+Incapacitated via488/487, cannotendownProne, laughteronlyifcapable; +1targetperslotabove1. Positive slot3threecreatures; reject4targets, intelligencefloorfrom2014, successconditions or ordinaryselfstand.',
  'repeat-ending': 'RepeatWISsaveendEACHtargetturn ANDeachdamagetaken; damagesaveAdvantage; successendsspell. Positive damagesaveuses2d20; reject damageautosuccess, missingturnsave, Advantageendturnsave or retainingboundconditionsafterending.',
 }),
 'Identify': (142, {
  'object': 'Level1;1minuteorRitual;Touchthroughout;VSM100GPpearlnotconsumed;instant. Learnmagicobjectproperties/use/attunement/charges, ongoingnamedspells andspellthatcreatedit. Positiveauthoritativecharge3returns3; reject briefonlytouch, pearlconsumed, fabricatedunknownitemeffects or grantingidentifiedeffectexecution.',
  'creature': 'Touchcreaturethroughoutcasting learnsongoingspellsONLY. PositiveactualongoingBlessreported; reject inventories, inherenttraits/secretcurseinference, unknowableprivatefacts or oneinstanttouchinsteadfullminute.',
 }),
 'Illusory Script': (142, {
  'write-view': 'Level1;1minuteorRitual;Touch;SM10GPinkconsumed;10days. Suitablewriting, caster/designatedreadnormalmeaning; othersunintelligiblescript ORalternatehand/meaning/knownlanguage. Positiveknownlanguagealternate; rejectunknownlanguage, laterarbitraryreaderchange, noinkcost.',
  'dispel-truesight': 'QualifiedDispelMagic makesoriginalANDillusiondisappear;Truesightreads hiddenmessage. PositiveactualTruesightread; reject deletingonlyillusion, universalprivateplaintextprojection, noqualifieddispeloutput or omissionbanon106dismissal.',
 }),
 'Imprisonment': (142, {
  'admission': 'Level9;1minute;30ftvisiblecreature;VSM5000GPstatuette(nonconsumed);WISsuccess unaffectedandimmunethisspell24h. Failedtargetnofood/drink/air/aging, cannotteleport, divinationcannotlocate/perceive. Positive24himmunity; reject consumingstatuette, ignoringpriorimmunity or successfulsaveimprisonment.',
  'burial': 'ChooseBurial: globejustlargeenoughtocontaintargetbeneathearth, nothingpassesin/out. Positiveboundedtargetcontained; reject creaturesbreathingneedrestored, unqualifiedterraincoordinates, arbitraryforceglobeattacks.',
  'chaining': 'ChooseChaining: groundrootedchains, Restrainedvia727andcannotbemovedANYmeans. Positiveforcedmovementdenied; reject onlySpeed0, ordinaryescapecheck or droppingbindingonrest.',
  'hedged-prison': 'ChooseHedgedPrison: demiplane labyrinth/cage/towerorlike, wardedteleportandplanartravel. Positiveboundednamedroomowner; reject arbitraryunqualifiedworldcreation, teleportescape or publicsecretcoordinates.',
  'minimus': 'ChooseMinimusContainment:target1inchtall trappedindestructiblegem/similar;lightpasses forseeingin/out,nothingelsepassesANYmeans. Positive1inch; reject1foot, opaquegem, breakablegem or inferredunrestrictedtelepathy.',
  'slumber': 'ChooseSlumber:Unconsciousvia487,cannotbeawoken. Positiveordinarydamagefailswake; reject rest/damageending, guessedconditionimmunities or genericUnconsciousremovalbypassingboundending.',
  'ending': 'AtcastspecifyobservableactiontriggerGMagreeslikelywithinnextdecade;unknown/unverifiableunavailable. DispelMagicONLYlevel9targetprisonORcomponent. Positivelevel9componentdispel; rejectlevel8checksuccess, unobservabletrigger, unilateralGMconsent or targetcreaturedispelwithoutsourceprisonidentity.',
 }),
 'Mind Spike': (149, {
  'damage': 'Warlockprinted75/header149consistent; disputedSorcererlistadmissionunavailable. Level2Action120ftvisiblecreature,S,Concentration1hour;WISsave3d8Psychic/halfsuccess,+1d8perslotabove2. PositivePact5six8s=48/24; reject slot5=3d8, successlocationgrant, unqualifiedSorcererlistcast.',
  'tracking': 'Failedsaveonly, whileconcentration/sameplane knowtargetlocation, cannotbehiddenfromcasterandInvisiblegivesnobenefitagainstcaster via962. PositivesameplaneInvisibleknownlocation; rejectcrossplanetracking, allplayerslocationleak, endingdamagewhenplanesdiverge or sightthroughobstructionsinferred.',
 }),
 'Mislead': (150, {
  'double-invisible': 'Level5ActionSelf,S,Concentration1hour;Invisiblevia962andillusorydoubleatownspace. Invisibilityendsafterownattackroll/dealdamage/castspell,doublepersists. PositivespellcastendsInvisibleonly; rejectretainingInvisible, destroyingdoubleonattack, physicaldoubleorindependentcreaturestats.',
  'move-perceive': 'MagicActionmovedoubleuptotwiceSpeedandgesture/speak/behave;intangible/invulnerable, see/hearthroughit. PositiveSpeed30move60; reject61ft, noMagicAction, attacks/physicalmanipulationorunknownprivateperceptiondisclosure.',
 }),
 'Planar Binding': (152, {
  'binding': 'Level5;1hour;60ftthroughout;VSM1000GPjewelconsumed;24hours. Celestial/Elemental/Fey/FiendCHAsavefailurebindservice; ifsummonedbyotherspellextendthatdurationtomatch. PositiveFeywithin60ftwholehour; rejectBeast,61ftinterval,successbound, concentrationwaiverinferred.',
  'commands-return': 'Boundfollowscommandsbestability;Hostilemaytwist. Completedtaskreports tocasterifsameplane,otherwise returnsbindingspotandremains. Positivecompletedtaskdifferentplanereturnsspot; rejectFriendlyconversion, omniscientcommands or mandatorysuicidewithoutsource.',
  'duration': 'Slot6=10days,7=30days,8=180days,9=366days. Positive9=366; reject365,slot8=30 or adoptingalladjacentsummonedcreatureeffects. Realnamedcreature/worldownersrequired.',
 }),
 'Ray of Enfeeblement': (157, {
  'success': 'Level2Action60ftcreature,VS,Concentration1minute;CONsaveSUCCESSnextattackDisadvantageuntilSTARTcasternextturn. Positive1nextattackdisadvantage; reject rangedspellattack2014, noeffectsuccess, multipleattackspenalizedorendnextturnexpiry.',
  'failure-ending': 'FailureSTR-basedD20TestsDisadvantageandsubtract1d8ALLdamagerolls; repeatCONsaveendtargetturn,successendsspell. Positivedamage12-d8=4->8andSTRsaveDisadvantage; rejectDEXtestdisadvantage, STRdamageonly, repeatstartturnorhalving2014.',
 }),
 'Scrying': (160, {
  'creature-save': 'Level5;10minutes;Self,VSM1000GPfocusnotconsumed;Concentration10minutes;sameplanecreatureWISsave modifiedknowledge secondhand+5/firsthand0/extensive-5,connectionlikeness-2/possession-4/bodypart-10. Positiveextensive+bodypart=-15; rejectguessedknowledge, consumedfocus, revealingprivatesavereason; successunaffected/castercannotrecasttarget24h.',
  'sensor': 'FailureInvisibleintangiblefistsizedsensorwithin10fttarget,followswithin10ft, castersees/hearsfromsensor; seeablesensorappearsluminousorb. Positive10ft; reject11ft, revealingcasterlocationfromsensor, physicalsensorattackorautomaticobserverdetection.',
  'location': 'Insteadtargetlocationyouhaveseen;stationarysensoratlocation,no creature-save branch. Positiveactualseenlocation; rejectheard-onlylocation,movinglocationsensororunknownworldcoordinates.',
 }),
 'Speak with Dead': (164, {
  'corpse-questions': 'Level3Action10ft,VSMburningincense,10minutes;corpsewithmouth, notUndeadatdeath, nottargetedpast10days;upto5questions,lifeknowledge/languagesonly. Positive5knownanswers; reject6,nomeaningfulmouth,9daypriorcast, futurepredictionorcompelledtruthenemy.',
  'knowledge-ending': 'Onlyanimatingspirit,nosoulreturn/resurrection/newlearning;briefcryptic/repetitiveanswerspossible. Positiveunknownpostdeathfactstaysunknown; rejectprivateworldanswersfabricated, newinformationlearnedbycorpse, resurrectiongrantorInvocationslotfreeeffectduplicate.',
 }),
 'True Polymorph': (170, {
  'admission-duration': 'Level9Action30ft,seencreatureornonmagicalobject,VSMmercury/gum/smoke,Concentration1hour;unwillingWISsave,successunaffected. Endsdeath/destruction;fullhourconcentrationbecomesuntildispelled. Positive60minpersistent; reject59minpersistent, worn/carriedobjecttargetortransformationendingat0THP2014.',
  'creature-form': 'Creature->creature CR<=targetCRorlevel;replacegamestatsbutretainHP/HitDice/alignment/personality;gainnewformHPasTHP,remainingTHPvanishspellend;nospeech/casting,anatomylimits. Positivelevel1target/currentHP20 with actualqualifiedOwlCR0/HP1 output retainsHP20 and gains1THP/Fly60/Flyby; rejectCR2form, replacingcurrentHP, retainedspellcastingorunqualifiednamedstatblock. Exactfinitepinnedworldroster routing independentofeffectavailability.',
  'equipment': 'Creature->creaturegear melds,newformcannotuseorbenefitfromequipment. Positivearmorbonusremovedwhilemeldedandrestoredending; rejectmagicitembenefits, orphanedgearorregrantingoldbonustwiceonreplay.',
  'object-creature': 'Object->creatureformsize<=objectsize,CR<=9,Friendlytoyou/allies,ownturnimmediatelyafteryours/obeyscommands; aftermorethan1hourcontrolends,mayremainFriendlydependingtreatment. PositiveagreedTinyglassbottle->actualqualifiedTinyOwlCR0/HP1/Fly60; rejectSmallformfromTinybottle,CR10orpermanentcommandcontrol. Namedstatblock/actions/lifecycleownersrequired; noundefinedform inferred.',
  'creature-object': 'Creatureandworn/carriedgear->objectsize<=creaturesize,statsbecomeobjectstats,norecollectionoftimewhenreturns. PositiveMediumcreature->knownagreedTinyglassbottleAC13/HP2; rejectLargeobjectfromMedium,keepingcreatureactionsorinventedgeneralobjectstats. Unbounddeath/overflow/endingoutputs unavailable; do notimport2014overflow/healing/gear duplication.',
 }),
 'Unseen Servant': (171, {
  'manifest-ending': 'Level1ActionorRitual60ft,VSMstring/wood,1hour;InvisiblemindlessshapelessMediumforceunoccupiedground;AC10/HP1/STR2,cannotattack;0HPendsspell. Positive1damageends; rejectAC12,attackpermission,occupiedspaceorTreatingforceasfamiliar.',
  'command': 'OnceeachownturnBonusActionmoveupto15ft/interactobject,simplehumantaskcontinuesuntilcompletewaitnextcommand; commandtaskthatwouldmove>60ftfromcasterendsspell. Positive15ftmoveat60ft; reject16ft,twocommandsperturn,61fttaskwithoutend,arbitraryqualifieditemeffectfrominteract.',
 }),
 'Vicious Mockery': (171, {'damage-disadvantage': 'Action60ft,V,instant;onecreatureseenORheardWISsavefailure1d6PsychicandnextattackDisadvantagebeforeenditsnextturn;characterlevels5/11/17=2/3/4d6. Positiveheardunseenlevel17four6s=24; rejectneitherseenheard, successdamage, allattacksDisadvantageorWarlocklevelscaling.'}),
 'Weird': (174, {
  'initial': 'Level9Action120ft,VS,Concentration1minute;chosencreatures30ftradiusSphereWISsave,failure10d10Psychic+Frightenedvia488;successhalfdamageonly. Positive10tens100/50success; reject31ftradius,Frightenedsuccess,2014nodamagefirstcastorunavailabletargetfacts.',
  'repeat-ending': 'FrightenedtargetWISsaveendeachownturn;failure5d10Psychic,successendsspellonthatTARGETonly. Positiveoneof2successends1/otherfails50; rejectalltargetsending,5d10success, repeatstartturnorFrightenedimmunitywithoutwhole488protocol.',
 }),
}

CASTING = ('Consume actual source-specific casting/target/dice/concentration outputs, not273/277 planning closure. '
           'Normal prepared Ritual-tag mode adds10minutes/no slot/no upcast; longcasts MagicActioneveryturn '
           'and castingConcentration/interruptionnoslotspend. Preserve106ongoingcasterNoActiondismissal '
           'unlessIncapacitated, source-specificstrongest/latest-equal same-spellnonstacking, exact206item/244scroll '
           'exceptions and confirmedprivateinvalidtargetslotspend/noeffect/noprivateleaks. '
           'These shared effects grant no class-list/feature entitlement; finite eligibility and unknown execution stay distinct.')


# Explicit audited whole-contract roster; no substring inference.
EXISTING_OWNERS = {'Acid Splash': (694,),
 'Alarm': (1035,),
 'Banishment': (939, 940),
 'Blight': (680,),
 'Burning Hands': (677,),
 'Command': (907,),
 'Detect Magic': (906,),
 'Detect Poison and Disease': (909,),
 'Dispel Magic': (355, 965, 966),
 'Druidcraft': (1049, 1050),
 'Elementalism': (1051, 1052, 1053),
 'Fire Bolt': (678,),
 'Fireball': (679,),
 'Geas': (945, 946),
 'Guidance': (954,),
 'Hold Person': (683,),
 'Insect Plague': (699,),
 'Jump': (1064,),
 'Light': (967,),
 'Magic Circle': (935, 936),
 'Mending': (955,),
 'Message': (1070,),
 'Misty Step': (688,),
 'Poison Spray': (1075,),
 'Power Word Kill': (506,),
 'Produce Flame': (1076, 1077),
 'Protection from Evil and Good': (913,),
 'Purify Food and Drink': (914,),
 'Ray of Frost': (684,),
 'Remove Curse': (937, 938),
 'Resistance': (956,),
 'Sacred Flame': (957,),
 'Shillelagh': (1079,),
 'Shocking Grasp': (689,),
 'Spare the Dying': (958,),
 'Speak with Animals': (700,),
 'Starry Wisp': (1085,),
 'Stinking Cloud': (697,),
 'Thaumaturgy': (959,),
 'Water Breathing': (1087,)}

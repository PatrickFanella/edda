"""Readable source execution expectations, transcribed independently of candidate bodies.

Each entry is a finite behavior owner. Protocol qualification does not implement it.
Source: pinned English SRD5.2.1; CC-BY4 attribution/adaptations in the source register.
"""
# name -> part -> (exact source fixture, concrete reused owners)
SPELLS = {
'Chill Touch': {'effect': ('PDF115 Action/Touch/VS/instant; melee spell attack target within reach,hit1d10Necrotic and cannot regainHP until ENDYOURNEXTturn; totallevels5/11/17->2/3/4d10. Positive:hitdamage+blockedhealing; negatives:old120ftrangedvariant,missrider,weaponattack,total4twoDice,permanenthealingblock.', [])},
'Circle of Death': {'effect': ('PDF115 level6 Action/150ft/VSM crushedblackpearlpowder500+GP notconsumed/instant;60ftradiusSphere,eachcreatureCONsave fail8d8Necrotic,successhalf;+2d8/slotabove6. Positive:slot7=10d8; negatives:DEXsave,slot7=9d8,151ft,unpaidmaterial,onlyenemies.', [])},
'Color Spray': {'effect': ('PDF116 level1 Action/self/VSM colorfulsand/instant;eachcreature15ftCone CONsave failBlinded untilENDYOURNEXTturn. Positive:failedCONadmit960 exactsource,end963 atsourceboundary; negatives:oldHPpoolvariant,DEXsave,successfulsaveBlinded,permanent,16ft.', [960,963])},
'Comprehend Languages': {'effect': ('PDF117 level1 Action ORRitual/self/VSM soot+salt/1hour; understandLITERAL heardspoken/seenSIGNED language; seenWRITTEN onlyTOUCHsurface,about1minute/page; no secretmessage/symboldecoding. Positive:touchpageplaintext; negatives:remoteunseenwriting,secretcipherdecoded,grantspokenfluency,ritualwithoutprepared/tag/+10minutes,ritualupcast.', [])},
'Dancing Lights': {'effect': ('PDF121 cantrip Action/120ft/VSM phosphorus/Concentration1min;uptoFOURtorchsize lights ORcombinefourintoONEvaguelyhumanlikeMediumform;eachDim10ft. BonusActionmovelightsupto60ftwithinrange;eachwithin20ftofanother sourceLight;exceedrangevanishes. Positive:movefour45ftwithBA; negatives:Brightlight,5lights,21ftseparation,freeActionmove,range121,lastingafterConcentration.', [])},
'Darkness': {'effect': ('PDF122 level2 Action/60ft/VM batfur+coal/Concentration10min;15ftradiusSpherefrompoint OR15ftEmanationfromobjectNOTworn/carried. OpaqueobjectcoverblocksDarkness;Darkvisioncannotsee,nonnmagicalLightcannotilluminate. OverlapANYareaBright/Dimfromspelllevel<=2 DISPELSthatotherSPELL. Positive:overlaplevel2Light; negatives:willingheldobjectadmission,Darkvisionbypass,level3lightdispel,coverendingratherthanblocking,nonspellLightdestroyed.', [])},
'Disguise Self': {'effect': ('PDF124 level1 Action/self/VS/1hour;self+onpersongear visualappearance,upto1ftshorter/taller,heavier/lighter,SAMEbasiclimbarrangement;physicalinspectionfailsillusion. StudyActionINTInvestigation vs YOURDC discern. Positive:1fttallerhumanoid; negatives:2ft,quadruped,realweight/statmutation,tactilesolidillusion,freeautomaticdisclosurewithoutphysicalinteraction/check.', [])},
'Expeditious Retreat': {'effect': ('PDF129 level1 BonusAction/self/VS/Concentration10min;takeDash780 NOW and mayDash asBonusAction UNTILends. Positive:initialDashandlaterBA; negatives:initialnoDash,freeDashwithoutBA,baseSpeedchange,repeatedBAoneTurn,afterend.', [780])},
'False Life': {'effect': ('PDF129 level1 Action/self/VSM alcohol/instant;gain2d4+4TEMPORARYHP,+5THP/slotabove1. Positive:injected2d4=5->9THP,slot2->14; negatives:old1d4,healingcurrentHP,stackTHPpools,Concentrationrequirement,FiendishVigormaxrollwithoutitsownentitlement.', [])},
'Feather Fall': {'effect': ('PDF130 level1 ReactiontriggerYOUorvisiblecreaturewithin60ftFALLS;60ft/VM feather/1minute;chooseuptoFIVEfallingcreatureswithinrange,descent60ft/ROUND. LandingBEFOREend causesNOfalldamage andendsforsaidcreature. Positive:fivefallerslandduringminute; negatives:6targets,nonfalling,61fttrigger,60ft/second,immunityafterexpiry,inventedwillingnessgate.', [])},
'Fire Storm': {'effect': ('PDF132 level7 Action/150ft/VS/instant;uptoTEN10ftCubes EACHcontiguousatleastoneother;eachcreatureDEXsave7d10Firefail/halfsuccess. Unworn/uncarriedflammableobjectsstartburning. Positive:connectedcubesandfailedsave; negatives:11cubes,disconnectedcube,doublecreaturedamageforoverlap,selectivelyspareallieswithoutsource,heldobjectignition.', [])},
'Greater Invisibility': {'effect': ('PDF137-138 level4 Action/Touch/VS/Concentration1min;ONEcreatureyoutouchInvisible962 UNTILspellends. Positive:attackdoesNOTendthisspellInvisible; negatives:ordinaryInvisibilityattackendingimported,remoteTouch,unrelatedInvisibledeleted,automaticinvisibilityagainstobserverwhocanseeyou. Use962observerAPI,notHide109xending.', [962])},
'Ice Knife': {'effect': ('PDF141 level1 Action/60ft/SM waterorice/instant;RANGEDspellattackONEcreature,hit1d10Piercing;HITORMISSexplodes,targetANDeachcreaturewithin5ftDEXsave fail2d6Cold,successZERO;ColdONLY+1d6/slotabove1. Positive:missstillColdexplosion; negatives:Vrequirement,missnoexplosion,savehalfdamage,upcastPiercing,excludeoriginaltarget.', [])},
'Knock': {'effect': ('PDF143 level2 Action/60ft/V/instant;VISIBLEaccessbarrierobject. Mundanelock/stuck/barredunlocked/unstuck/unbarred;ONLYONEifmultiplelocks. ArcaneLock SUPPRESSED10minutes,notdispelled. Loudknockaudible300ftemanatesfromTARGET. Positive:onelockopenwithsound; negatives:alllocks,ArcaneLockpermanentending,silentcastwithoutsound,soundfromcaster,unseenobject. Rogue lockfixtures ownmundane checks,notthisspell.', [])},
'Legend Lore': {'effect': ('PDF144 level5 10MINUTES/self/VSM incense250+GP CONSUMED plusFOURivorystrips50+GPEACH notconsumed/instant. Name/describefamousperson/place/object;GM-explicitaccuratelore,precisionfromactualpriorinfo;mayfigurative. Notfamous->sadtrombonenotesandFAIL. Positive:authorizedfamousitemsummary; negatives:inventedprivatelore,4strips50GPtotal,unconsumedincense,Actioncast,inferredfame. Consume105longcastingprotocol,privateoutcomefilter,slotspentonnoneffectasapplicable.', [])},
'Levitate': {'effect': ('PDF144-145 level2 Action/60ft/VSM metalspring/Concentration10min;ONEvisiblecreatureORlooseobject<=500lb,riseverticallyupto20ft. UnwillingcreatureCONsavesuccessunaffected. Movementonlypush/pullfixedsurfaceasclimb. OnYOURturn changealtitudeupto20ft;SELFpartofmove,OTHERtargetMagicActionandwithinrange. Endingfloatgentlyground. Positive:selfmove20ft; negatives:501lbobject,unwillingautosuccess,otherfreeaction,horizontalfreeflight,endingfalldamage.', [])},
'Mage Armor': {'effect': ('PDF145 level1 Action/Touch/VSM curedleather/8hours;WILLINGunarmoredcreature,baseAC13+DEXmodifier;donningarmorENDSspell. Positive:DEX2->15; negatives:armored,unwilling,add13toexistingAC,DEXscoreinsteadmodifier,armorremovedresumesalreadyendedspell. Nofocuspossessionalonegrantscasting.', [])},
'Magic Missile': {'effect': ('PDF146 level1 Action/120ft/VS/instant;THREEdarts,eachVISIBLEcreaturechosen,1d4+1ForceEACH,simultaneous,oneorseveraltargets;+ONEdart/slotabove1. Positive:slot2fourdartsexactallocations; negatives:attackroll,save,objecttarget,unseen,upcastdamageinsteadcount,sequentialdeathretargeting. RespectexactShieldspell immunitywithoutimportingclasseligibility.', [])},
'Meteor Swarm': {'effect': ('PDF149 level9 Action/1MILE/VS/instant;FOURdifferentVISIBLEpoints,each40ftradiusSphere;creatureDEXsave fail20d6FireAND20d6Bludgeoning,successhalf;overlapcreatureaffectedONLYONCE. Nonmagicunworn/uncarriedobjectsalsotakedamageandflammablestartburning. Positive:overlappedcreatureonecombineddamage; negatives:fourtimesoverlap,unseenpoint,5points,magicitemdamagedbyobjectclause,ignoreresistanceseparatetypes.', [])},
'Scorching Ray': {'effect': ('PDF159 level2 Action/120ft/VS/instant;THREEraysoneorseveraltargets,SEPARATErangedspellattackEACH;hit2d6Fire;+ONEray/slotabove2. Positive:slot3fourattackrolls,oneMissnodamage; negatives:oneattackallrays,upcastdamageinsteadcount,automaticHit,121ft,unsavedrerollonretry.', [])},
'Shatter': {'effect': ('PDF161 level2 Action/60ft/VSM mica/instant;10ftradiusSphere,eachcreatureCONsave fail3d8Thunder/halfsuccess;CONSTRUCTdisadvantageonsave. Nonmagicalunworn/uncarriedobjectsalsoTAKEdamage;+1d8/slotabove2. Positive:slot3=4d8,Constructdisadvantage; negatives:DEXsave,allobjectsimmune,heldobjectdamage,magicobjectdamage,constructautosavefailure.', [])},
'Shield': {'effect': ('PDF161 level1 REACTIONwhenHITbyattackroll ORtargetedbyMagicMissile;self/VS/1round UNTILstartYOURNEXTturn,+5AC INCLUDINGtriggerattack andNOMagicMissiledamage. Positive:hitAC15roll18becomesmissAC20; negatives:afterresolveddamagewindow,ordinarysaveeffecttrigger,allytarget,endsendturn,onlytriggeringattackbenefit,noV/S. Preserveoutcome-independentknowndiewindows,nonewinventedunknownoutcomeban.', [])},
'Silent Image': {'effect': ('PDF162 level1 Action/60ft/VSM fleece/Concentration10min;visualONLYobject/creature/phenomenon<=15ftCube;MagicActionmovewithinrangeandnaturallyalterappearance. Physicalinteractionrevealsillusion;StudyINTInvestigationvsDCdiscerns,discerningcreatureseesthrough. Positive:moveimagewithMagicAction; negatives:sound/smell,tactilebarrier,16ftCube,freeMove,automaticomniscientNPCdisclosure.', [])},
'Spider Climb': {'effect': ('PDF164 level2 Action/Touch/VSM bitumen+spider/Concentration1hour;ONEwillingcreaturemovesverticalsurfaces/ceilingsHANDSfree,ClimbSpeed=Speed;+ONEcreature/slotabove2. Positive:slot3twowillingclimbers; negatives:unwilling,freeFly,twiceSpeed,slot2twotargets,losehandsfree. DeriveSpeedauthoritatively,notuntrustedclientvalue.', [])},
'Thunderwave': {'effect': ('PDF169-170 level1 Action/self/VS/instant;15ftCubeFROMcaster,eachcreatureCONsave fail2d8ThunderANDpush10ftAWAY,successHALFdamageONLY. UnsecuredobjectsENTIRELYwithinCubealsopush10ft;boom300ft. +1d8/slotabove1. Positive:failedsavepush,savesuccessnopush; negatives:DEXsave,spheregeometry,partialobjectpush,objectsecured,boomfromtargetinsteadsource.', [])},
'Tongues': {'effect': ('PDF169 level3 Action/Touch/VM miniatureziggurat/1hour;creatureunderstandsHEARDspoken/SEENsignedlanguage;itsspeech/signingunderstoodbycreatureknowing>=ONE language thatcanHEAR/SEEit. Positive:seenforeignsign; negatives:readwrittenlanguage,unheard/unseensignal,recipientknowsNOlanguage,omnipresenttranslation,forcedwillingpredicate.', [])},
'Vitriolic Sphere': {'effect': ('PDF172 level4 Action/150ft/VSM bile/instant;20ftradiusSphereDEXsave;FAIL10d4AcidNOW+5d4AcidENDtargetNEXTturn;SUCCESSHALFinitialONLY. InitialONLY+2d4/slotabove4. Positive:slot5fail12d4then5d4,savesuccessnodelayed; negatives:successdelayed,upcastdelayed7d4,CONsave,duplicatednextturntick,151ft.', [])},
}

SPELLS.update({
'Alter Self': {
 'appearance':('PDF107 level2 Action/self/VS/Concentration1h;chooseONEphysicalformoption;MagicActionreplacesoption. Appearance maychangeheight/weight/voice/features/species butSAMESIZE/basicshape andNOstatisticchange;MagicActionchangeappearanceagain. Positive:humanappearancewithoutstatgrant; negatives:quadrupedbipedchange,sizechange,newancestryfeatures,keepotheroptionbenefit,freechange.',[]),
 'aquatic':('PDF107 ownactiveAlterSelf Aquaticoption;underwaterbreathing+SwimSpeed=Speed. Replacingoptionendsownedbenefits. Positive:Speed30Swim30; negatives:Swim60,waterbreathingafteroptionchange,freeFly,inferredoxygenfacts. Consumeappearanceoptionselector,notnewcast.',[]),
 'natural-weapons':('PDF107 ownactiveAlterSelf NaturalWeapons chooseclawsSlashing/fangsPiercing/hornsPiercing/hoovesBludgeoning;UnarmedStrike779 DAMAGEwithgrowth uses1d6 insteadnormal andcastingabilitymodifier attack/damage insteadSTR. Positive:CHA3clawsD6=4->7Slashing; negatives:extraD6onoldnormaldamage,grapplediebonus,wrongtype,STRforced,benefitafteroptionreplacement.',[779]),
},
'Animate Objects': {
 'summon-scaling':('PDF108 level5 Action/120ft/VS/Concentration1min;nonmagicunworn/uncarried/unfixed NONgargantuanobjects. Budget=castingabilityMODIFIER;Mediumorsmaller1,Large2,Huge3. BecomesConstruct:AC15,HP10/20/40,sizebased,Speed30,STR16/DEX10/CON10/INT3/WIS3/CHA1,PB=yourPB;Poison/Psychicdamageimmune andCharmed/Exhaustion/Frightened/Paralyzed/Poisonedimmune;Blindsight30,Passive6,understandcasterlanguages. Positive:CHA3oneHuge; negatives:CHA3twoLarge,Gargantuan,magic/held/fixedobject,old10objecttable,missingexactstatprovenance.',[184,487,488,727]),
 'initiative-commands':('PDF108 alliedcaster/allies,sharescasterInitiative,turnimmediatelyAFTERcaster. BonusActionMENTALcommandownedobjectswithin500ft;multipleSAMEcommandperBA. Nocommand=Dodge782+moveONLYavoidharm. Positive:commandtwotogether; negatives:501ft,newfreeaction,verbalrequirement,separateInitiative,autonomousattack,distinctcommandsoneBA.',[782]),
 'actions':('PDF108 ownedAnimatedObject SlamMELEE attackYOURspellattackmodifier reach5,FORCEdamage:Mediumorsmaller1d4+3;Large2d6+3+castingabilityMOD;Huge2d12+3+MOD. Eachslotabove5addsD4/D6/D12byrespectivesize. Positive:LargeCHA3slot6=3d6+6; negatives:physicalPiercingtype,allsizeD4,missingCHAforLarge,addCHAtoSmall,attackoutside5ft,unknownstatblockaction.',[]),
 'ending':('PDF108 controlledUNTILspellendsOR0HP. At0HP revertsobjectform andREMAININGdamagecarryovertoobject. Positive:5HPobjecttakes8->3damageobjectform; negatives:0HPpermanentConstruct,nooverflow,full8again,foreignobjectremoval,restoredmagicalitem. Spellendingseparatefrom0HPformtransition;removeonlysourceownedworldinstance.',[]),
},
'Arcane Eye': {
 'sensor':('PDF110 level4 Action/30ft/VSMbatfur/Concentration1h;InvisibleINVULNERABLEhoveringeye,mentalVISUALinformationinALLdirections,Darkvision30. Positive:authorizedvisiblefactsaroundeye; negatives:hearing,Darkvision60,destructibleeye,omniscienthiddenfacts,31ftinitialplacement. Sensepermissionsandperceptiblefactsremainauthoritative.',[962]),
 'movement-disclosure':('PDF110 activeownArcaneEye BONUSActionmoveupto30ftANYdirection;SOLIDbarrierblocks,opening>=1inchpasses. Positive:1inchhole; negatives:0.5inch,solidwall,31ftperBA,freeMagicAction,sourceundocumentedmaxdistancefromcasterimposed. Returnonlysourcepermittedobservations.',[]),
},
'Arcane Hand': {
 'manifest-scaling':('PDF110 level5 Action/120ft/VSMeggshell+glove/Concentration1min;LargeMAGICALOBJECTinVISIBLEUNOCCUPIEDspace,AC20,HP=yourMAXHP;doesNOToccupyspace. InitialcastandlaterBONUSActionmoveupto60ftTHENonequalifiedoption. Positive:castermax40handHP40; negatives:currentHPinsteadmax,creaturestatblock,occupied/unseenplacement,handblocksallspace,free60ftmoveeveryevent.',[]),
 'clenched-fist':('PDF110 ownhandoption;targetwithin5ftMELEE spellattack,hit5d8Force;+2d8/slotabove5. Positive:slot6=7d8; negatives:reach6,missdamage,old4d8,rangedattack,unqualifiedhand. Initialcast/laterBAoptionprotocolonly.',[]),
 'forceful-hand':('PDF110 Hugeorsmallercreaturewithin5fthand,STRsavefailpushupto5+5*castingabilityMODfeet;handmoveswithtargetremainingwithin5ft. Positive:CHA3push20; negatives:Gargantuan,DEXsave,pushonsavesuccess,handleftbehind,usingabilityscore,push21.',[]),
 'grasping-hand':('PDF111 Hugeorsmallercreaturewithin5fthandDEXsavefailGrappled778 escapeDC=yourspellDC. WhilegrappledBONUSActioncrush4d6+castingabilityMODBludgeoning;+2d6/slotabove5. Positive:slot6CHA3=6d6+3; negatives:STRadmissionsave,crushwithoutgrapple,freecrush,wrongescapeDC,ForceinsteadBludgeoning,copygenericunarmedcost.',[778]),
 'interposing-hand':('PDF111 ownhandHalfCoveragainstATTACKS/othereffectsORIGINATINGinhandspace ORpassingTHROUGH;spaceDifficultTerrainforENEMIES. Positive:crossingrayHalfCover; negatives:allattacksregardlesspath,ThreeQuartersCover,alliesDifficultTerrain,solidwallpassageban. Authoritativegeometry/factionfactsrequired.',[]),
 'ending':('PDF110 hand0HPENDSspell;Concentration/time1min/own106dismissalendhand,removeownedcover/terrain/grappleeffects. Positive:0HPendsallownedcontributionsonce; negatives:disappearbutkeepgrapple,clearunrelatedGrappled,0HPdoesnotendspell,Incapacitated487dismissal,foreigncasterending.',[487,778]),
},
'Blindness/Deafness': {
 'admission':('PDF113 level2 Action/120ft/V/1minNOConcentration;VISIBLEcreatureCONsavefailchooseBlinded960ORDeafened961;targets=slotlevel-1. Positive:slot3twoindependentsaves; negatives:60ftoldrange,VSrequirement,Concentration,bothtypes,successfulsavecondition,3targetsat3.',[960,961]),
 'repeat-save-ending':('PDF113 eachaffectedtargetENDitsEACHturnrepeatsCONsave,successendsONLYownspelltarget;timeout1minalsoends. Consume963/964exactsourceendings. Positive:twotargetsonesuccessotherremains; negatives:casterturnsave,removeotherBlindedsource,requiresAction,savestopsafterfirstfailure.',[963,964]),
},
'Blink': {
 'roll-plane-transition':('PDF113 level3 Action/self/VS/1minNOConcentration;ENDyourEACHturnrollD6,4-6Etherealtransition;alreadyEthereal->spellENDS. Thereonlyaffect/affectedbyEtherealcreatures;oldplaneseenGRAYupto60ft,otherscannotperceiveexceptactualspecialsense. Positive:D6=4transition; negatives:D20oldvariant,3success,Concentration,damageotherplane,61ftvision,privatefactleak.',[]),
 'return-ending':('PDF113 whenEthereal returnSTARTnextturn ANDspellending;chooseVISIBLEunoccupiedspacewithin10ftleftspace,ifNONEnearestunoccupiedspace. Positive:choose10ft; negatives:11ftwhenlegalspaceavailable,occupied,randomteleport,stayEtherealafterend,unverifiednearestspace. Persistplane/position/turn,notworldmodelguess.',[]),
},
'Chain Lightning': {'targets-damage':('PDF114 level6 Action/150ft/VSM3silverpins/instant;VISIBLEfirstcreatureORobject,uptoTHREEOTHERtargetsEACHwithin30ftFIRST,eachonlyONEbolt. EACHDEXsave10d8Lightningfail/halfsuccess;slotabove6+ONEleapingbolt. Positive:slot7fiveuniquetargets; negatives:chain30ftfromlastinsteadfirst,duplicatebolt,target151ftfirst,saveCON,upcastdamageinsteadcount.',[])},
'Charm Monster': {
 'charm':('PDF114 level4 Action/30ft/VS/1hourNOConcentration;ONEVISIBLEcreatureWISsave,AdvantageifcasterORalliesfightingit;failCharmed488+FriendlyTOCASTER. +ONEtarget/slotabove4. Positive:nonHumanoidfail; negatives:Humanoidonly,missingcombatAdvantage,friendlytoallallies,autocontrol,unseen/31ft.',[488]),
 'ending-awareness':('PDF114 ownCharmedENDSspellendORcaster/ALLIESdamageit;whenENDS,targetKNOWScharmedBYcaster. Positive:allydamageendsanddisclosesidentityonlytarget; negatives:anyenemyDamageend,disclosebeforeending,clearotherCharm,omitawareness. Trusted488endingsonly.',[488]),
},
'Charm Person': {
 'charm':('PDF115 level1 Action/30ft/VS/1hourNOConcentration;VISIBLEHUMANOIDWISsave,AdvantageifcasterORalliesfightingit;failCharmed488+FriendlyTOCASTER;slotLtargetsL. Positive:slot2twoHumanoids; negatives:Beast,missingcombatAdvantage,CharmonSaveSuccess,mindcontrol,31ft.',[488]),
 'ending-awareness':('PDF115 ownCharmendswhenSPELLendsORcaster/ALLIESdamageit;targetthenKNOWScharmedBYcaster. Positive:casterdamageends; negatives:enemyonlydamageending,noawareness,clearunrelatedCharm,publicdisclosureofprivatetargetidentity. Consume488exactsourceending.',[488]),
},
'Chromatic Orb': {
 'damage':('PDF115 level1 Action/90ft/VSMdiamond50+GPnotconsumed/instant;targetrangedspellattackhit3d8chooseAcid/Cold/Fire/Lightning/Poison/Thunder;+1d8/slotabove1. Positive:slot2Fire4d8; negatives:Psychic,consumeddiamond,49GP,hitwithoutattack,upcasttargetcountwithoutbouncepredicate.',[]),
 'bounce':('PDF115-116 acceptedhitdamaged8swith>=TWOEQUALnumbers mayleapDIFFERENTtargetwithin30ftpriorTARGET;NEWattack+NEWdamage. Slot1maxONEleap;slotLmaxLleaps;creaturetargetedonlyONCEperCAST. Positive:3/3/5damageleap,slot2upto2leaps; negatives:noequalDice,31ft,repeatcreature,reusedattack/damage,oldnocanbouncevariant,infiniteleaps. Persistorderedtargets/dice/cap.',[]),
},
'Clairvoyance': {
 'sensor':('PDF115-116 level3 10min/1MILE/VSM100+GPfocus,hornhearingORglasseyeseeing/Concentration10min;locationFAMILIARvisited/seenOROBVIOUSunfamiliar;InvisibleINTANGIBLEINVULNERABLEIMMOBILEsensor. Positive:knownbehinddoorlocation; negatives:arbitraryunfamiliarnonobviouslocation,99GP,Actioncast,movingsensor,inferredmap/privatefacts.105longcastprotocol.',[962]),
 'sense-choice-disclosure':('PDF116 chooseSEEINGorHEARINGatCAST,throughsensorasifthere;BONUSActionswitch. CreatureACTUALLYseeingsensorviaqualifiedsenseperceivesfistsizeluminousorb. Positive:BAchangesseeingtohearing; negatives:bothsimultaneouslywithoutchoice,freechange,ordinaryobserverseesInvisibleorb,unqualifiedTruesightentitlement.',[962]),
},
'Confusion': {
 'admission':('PDF117 level4 Action/90ft/VSM3nutshells/Concentration1min;EACHcreature10ftradiusSphereWISsavefailNOBonusActions/Reactionandturntable. +5ftradius/slotabove4. Positive:slot5radius15; negatives:radius10forever,onlyenemy,CONsave,successfulsaveaffected,missingReactionban.',[]),
 'turn-table':('PDF117 affectedtargetSTARTturnD10:1NOAction,allmovementdirectionD4=1N/2E/3S/4W;2-6NOmove/actions;7-8NOmove,AttackONEmeleeattackRANDOMcreaturewithinreach,nonethenNOAction;9-10targetCHOOSESbehavior.BonusAction/Reactionbanremains. Positive:7randomknownreachfixture; negatives:casterchoosesvictim,2moves,9stillrandom,untrustedhiddenactorselection. SaveallDice/choice/turn.',[]),
 'ending':('PDF117 ENDaffectedtargetEACHturnWISsave,successendsownrecipient;Concentration/timeendexactsourceinstances. Positive:onerecipientends; negatives:startturnsave,noRepeats,unrelatedDebuffcleared,expiredtableeventexecuted.',[]),
},
'Counterspell': {
 'player-and-general-casting-trigger':('PDF120 level3 REACTIONwhenYOUSEEcreaturewithin60ftCASTINGspellwithV ORS ORM;S/instant. Extend356exactpendingattemptidentity/revisionprotocoltoqualifiedplayer/generalsourcecastoutputsWITHOUTrewriting356NPC/CureWoundsfixture. Positive:observedSonlycast; negatives:61ft,unseen,noactualcomponentsSubtlecasting,afterresolution,knownspellnameforceddisclosure,castingabilitycheckoldvariant.',[356]),
 'save-cost-resume':('PDF120 acceptedpendingcast:TARGETCONsavevsreactorDC;FAILspellNOeffectandcastingAction/BA/ReactionWASTED,spellslotIFUSEDNOTexpended;SUCCESSoriginalcastresumesONCE. ReactorReaction+slotspentonce. Positive:failedCONretainsoriginalslot; negatives:targetslotspent,refundtargetaction,successcancels,rerollonreload,atomicityonlyoneactor. Coordinateexact356transactionandoriginalspellowner;notgenericNPCcastingadmission.',[356]),
},
'Creation': {
 'object-material':('PDF121 level5 1MINUTE/30ft/VSMpaintbrush/special;ONEobject<=5ftCube,vegetableORmineralmatter,FORMandMATERIALyouhaveSEEN. +5ftCubeside/slotabove5. Positive:slot6seenwood10ftCube; negatives:livingcreature,6ftat5,unseenform/material,inventedvaluablemagicitem,Actioncast.105longcastprotocol.',[]),
 'duration-ending':('PDF121 materialdurationvegetable24h,stone/crystal12h,preciousmetal1h,gems10min,adamantine/mithral1min;MIXEDusesSHORTEST. UsingcreatedobjectasANYotherspellMATERIALcausesTHATspellFAIL. Positive:wood+gold1h; negatives:longest24h,materialsubstitutionpasses,consumednormalmaterialinsteadfailedspell,permanentobject,destructionunrelatedobjects.',[]),
},
'Cloudkill': {
 'area-movement':('PDF116 level5 Action/120ft/VS/Concentration10min;20ftradiusSphereHeavilyObscuredfog;STARTeachYOURturnmoves10ftAWAYfromyou;STRONGwinddispersesENDSspell. Positive:authoritative10ftmove; negatives:stationaryoptional,moveANYdirection,ordinarybreezeending,untrustedwind/worldfacts,afterendfogleftactive.',[1061]),
 'damage-events':('PDF116 initialEACHcreatureSphereCONsave5d8Poisonfail/halfsuccess;alsoSphereMOVESINTOspace,targetENTERSorENDSturthere;ONLYONCEperTURNpercreature.+1d8/slotabove5. Positive:initialthenendSAMEturnoneevent; negatives:immunePoisonbypass,onceROUNDinsteadturn,standingSTARTturntick,duplicatehistory,slot6still5d8.',[]),
},
'Disintegrate': {
 'creature':('PDF124 level6 Action/60ft/VSMlodestone+dust/instant;VISIBLEcreatureDEXsaveFAIL10d6+40Force,+3d6/slotabove6;SUCCESSnoeffect. IfTHISdamagereduces0HP creatureANDnonmagicworn/carriedgeardust,revivalONLYTrueResurrection/Wish. Positive:slot7=13d6+40; negatives:halfonsuccess,any0HPbeforethisspell,magicgeardusted,ordinaryrevival,noactualqualifieddeathreceipt.',[]),
 'object-force':('PDF124 samecastVISIBLEnonmagicOBJECTORcreationMAGICALFORCE;LargeorsmallerAUTOMATICdisintegration,Hugeorlarger10ftCubePORTION. Positive:magicalForcewallportion; negatives:magicitem,nonsourceenergywall,entireHugeobject,creaturesautosuccess/noSave,unknownmaterial/ownership. Actualboundedworlddestructionoutputrequired.',[]),
},
'Fear': {
 'admission':('PDF130 level3 Action/self/VSMwhitefeather/Concentration1min;EACHcreature30ftConeWISsavefailDROPeverythingheldandFrightened488. Positive:failedWISdropgear; negatives:successdrop,creatureoutsideCone,CONsave,permanentitemdeletion,selectivelysparewithoutsource.',[488]),
 'movement-ending':('PDF130 affectedcreatureEACHturnDASH780andmovesAWAYbysafestROUTEunlessNOWHEREmove. ENDturnWITHOUTLOS toCASTER WISsave,successendsownspellrecipient. Positive:occludedendturnsave; negatives:savewithLOS,unsafechosenroute,inventedautomaticGMpath,freeendingAction,clearallFrightenedinstances.',[488,780]),
},
'Finger of Death': {
 'damage':('PDF131 level7 Action/60ft/VS/instant;VISIBLEcreatureCONsave7d8+30Necroticfail/halfsuccess. Positive:failedsaveexactsourceDamage; negatives:DEXsave,unseen,61ft,automatickillthreshold,oldspellconstitutioncheck,inferredprivateHPleak.',[]),
 'zombie-creation-control':('PDF131 HUMANOIDKILLEDbyTHISspell risesSTARTYOURNEXTturnasZOMBIEobeyingYOURVERBALorders. ConsumeactualsourceZombie178dataPLUSownboundedZombiestat/actionexecution;not185planningorstockNPC. Positive:humanoiddeathspawnsoneownedZombieatboundary; negatives:Beast,notkilled,anotherdamagecause,immediatespawn,mental/freeactorcommands,duplicateworldentity. Sourcecreatedcreaturehasnoinventedconcentrationduration.',[178]),
},
'Sorcerous Burst': {
 'attack':('PDF163 cantrip Action/120ft/VS/instant;ONEcreatureORobjectRANGEDspellattackhit1d8choiceAcid/Cold/Fire/Lightning/Poison/PSYCHIC/Thunder;TOTALlevels5/11/17->2/3/4d8. Positive:Psychichit; negatives:Radiant/Force,melee,missDamage,classlevelscaling,total4twoDice.',[]),
 'exploding-dice':('PDF163 wheneveran8rolledonD8forthisSPELL mayrollADDITIONALD8andadddamage;TOTALadditionalDicecap=castingabilityMODIFIERforCAST. Positive:CHA2,initial8then8permits2extraMAX; negatives:uncappedchain,capPB,8onotherdie,rerollold8insteadaddition,perhitresetcap,musthaveknownoutcomedisclosurebeforewindow. Persistentirediechain.',[]),
},
'Suggestion': {
 'suggestion':('PDF166 level2 Action/30ft/VMhoney/Concentration8h;ONEVISIBLEcreatureHEARS/UNDERSTANDSsuggestion<=25words,ACHIEVABLEandnotobviousdamagetargetORallies;WISsavefailCharmed488andpursuebestability. Positive:boundedtrustedfetchkeycommand; negatives:26words,unheard,unachievable,obviousallyharm,freeformmodelactioninference,oldcreaturetypeimmunityinvented.',[488]),
 'duration-ending':('PDF166 ownCharmedendsSPELLend ORcaster/ALLIESdamagetarget ORACTIVITYCOMPLETEevenbefore8h. Positive:activitycompleteend; negatives:enemyDamageautoends,obviousharmallowedaftercast,lastingaftercompletion,clearotherCharm,castingrequiresordinaryFriendly.',[488]),
},
})

SPELLS.update({
'Delayed Blast Fireball': {
 'bead-growth':('PDF122 level7 Action/150ft/VSMguano+sulfur/Concentration1min;chosenpointglowingbead,BASE12d6Fire,+1d6EACHYOURturnENDwhileongoing;+1d6BASE/slotabove7. Positive:slot8after2turnends15d6; negatives:growonotherturn,startturngrowth,12fixedonexpiry,immediatedamageatcast.',[]),
 'touch-detonation':('PDF122 touchBEFOREspellend->touchingcreatureDEXsaveFAILENDSspell/explode;SUCCESSmayTHROWupto40ft;thrownbeadenterscreaturespaceORsolidcollisionENDS/explodes. Positive:success40ftthrow; negatives:saveCON,failedsavethrow,41ft,untrustedcollisionfacts,damagebeforetouchdecision.',[]),
 'ending-damage':('PDF122 WHENSPELLENDSbeadexplodes20ftradiusSphere,EACHcreatureDEXsave accumulatedFirefail/halfsuccess;unworn/uncarriedflammableobjectsignite. Positive:concentrationbreakdetonatesonce; negatives:endwithoutblast,damageunselectedwholemap,multipleblastsonretry,heldobjectignition. Consumeownedaccumulateddiecountandactualendingevent.',[]),
},
'Demiplane': {
 'door':('PDF122 level8 Action/60ft/S/1hNOConcentration;shadowyMEDIUMdooronVISIBLEFLATSOLIDsurface;open/closeleadsemptyroom30ftEACHdimensionwoodORstone. Enddoorvanishes;objectsREMAIN,creaturesREMAINunlessOPTshuntedthroughvanishingdoor,Prone488nearestunoccupiedspaces. Positive:optinshuntProne; negatives:automaticallcreatureshunted,objectsreturned,31ftroom,VSrequired,unverifiedspace.',[488]),
 'persistent-room-access':('PDF122 eachcastNEWdemiplaneOROWNpreviousone;OTHERSroomONLYifKNOWSnatureANDcontents. Positive:ownroomreopenpersistedcontents; negatives:knownameonlyotherroom,guessprivatecontents,anyplaneportal,deletecontentsonexpiry. Authority/roomidentity/knowledgeprovenancepersisted,notloreinference.',[]),
},
'Detect Thoughts': {
 'detection-ending':('PDF123 level2 Action/self/VSM1CopperPiece/Concentration1min;initialSenseThoughtsORReadThoughts,laterMagicActionEITHER. Sensepresencewithin30ftcreaturesknowLANGUAGEorTELEPATHIC,notthoughtcontent;blocked1ftstone/dirt/wood,1inchmetal,thinlead. Positive:telepathicpresence; negatives:noLanguage/notelepathy,31ft,contentdisclosedbysense,1inchwoodblockinstead1ft. Sourceendingremovessensorsonly.',[]),
 'surface':('PDF123 ReadThoughtsONEVISIBLEcreaturewithin30ftORoneDETECTEDbyownSenseThoughtswithin30ft;learnMOSTonmindNOW;nolanguage/notelepathy->NOTHING. Positive:previouslydetectedsamecreaturebehindcoverwhereauthorized; negatives:randomunseennondetectedtarget,oldermemoryasautomaticoutput,31ft,GMprivatefactsinferred. GM-authorizedthoughtreceipt/disclosureonly.',[]),
 'probe':('PDF123 MagicActiononNEXTturnprobeSAMEtargetdeeper->WISsaveFAILreasoning/emotions/loomingmatter,SUCCESSSPELLENDS;EITHERWAYtargetKNOWSprobe. Untilattentionshift targetmayACTIONINTArcana vsYOURDC successENDSSPELL. Positive:successWISend+awareness; negatives:probeinitialturn,noawarenessonfailure,freeArcanacheck,indefiniteprobingafterattentionchange,AIinventedthoughts.',[]),
},
'Dimension Door': {
 'transport':('PDF123 level4 Action/500ft/V/instant;destinationseenORvisualizedORdescribeddistance/direction;exactspot. MaytakeONEwillingcreaturewithin5ft,arriveswithin5ftdestination. Positive:unseenexplicitmapped200ftdown; negatives:501ft,2companions,unwilling,companion6ftaway,blanketseendestinationgate,unknownauthoritativepositioninferred.',[]),
 'blocked-arrival':('PDF123 ifEITHERarrivaloccupiedbyCREATUREORcompletelyfilledobjects,TELEPORTFAILS andcasterANDcompanionEACH4d6Force. Positive:blockedcompanionarrivaldamagesbothandneithermoves; negatives:nearestfreespacefallback,onlycompaniondamage,partialobjectcountscompletelyfilled,inferredsecretobstructiondisclosure. Confirmcost/noeffectprivacyandatomicbothactorpositions.',[]),
},
'Dominate Monster': {
 'charm':('PDF125 level8 Action/60ft/VS/Concentration1h;ONEVISIBLECREATUREWISsavewithAdvantageifcaster/alliesfighting;failCharmed488. Slot9duration8h. Positive:BeastorHumanoideligible; negatives:Humanoidonly,initial1min,missingcombatAdvantage,31ftinventedrange,controlbeforeresolvedsave.',[488]),
 'commands':('PDF125 acceptedownCharm SAMEPLANEtelepathiclink;YOURturncommandsNOaction,targetobeybestonOWNturn;completedorderwithoutnewdirectionfreebehavior/selfprotection. CommandReactionrequiresBOTHYOURandTARGETReaction. Positive:acceptedboundedcommand; negatives:otherplanetelepathy,commandanyturn,onlytargetReactioncost,automatictotalmindpossession,modelinferredaction. No1048copy:thatownerisDominateBeastonly.',[488]),
 'ending':('PDF125 EACHdamageeventtargetREPEATSWISsave,successendsownrecipient. Spell/Concentration/timeendingendsownCharm/controlonly. Positive:enemydamagetriggerssave; negatives:onlycasterdamage,saveonceperround,allCharmscleared,telepathiccommandsafterending.',[488]),
},
'Dominate Person': {
 'charm':('PDF126 level5 Action/60ft/VS/Concentration1min;ONEVISIBLEHUMANOIDWISsave,Advantageifcaster/alliesfighting;failCharmed488. Slot6=10min,7=1h,8+=8h. Positive:slot7Humanoid1h; negatives:Beast,8hatslot6,missingcombatAdvantage,controlwithoutfailedsave.',[488]),
 'commands':('PDF126 ownCharm SAMEPLANEtelepathy,YOURturnNOactioncommand,bestobeyonTARGETturn;completedorder/nootherselfprotect/freebehavior. ReactioncommandspendsBOTHcaster+targetReaction. Positive:boundedsavedAttackcommand; negatives:totalactiontakeover,otherplane,onlycasterReaction,noqualificationofcommandedaction,modelinferredenemyorders.',[488]),
 'ending':('PDF126 EVERYtargetdamageREPEATWIS,successendsownspellrecipient;sourceConcentration/timeendingendscontrol. Positive:allydamageandthirdpartydamagebothsaves; negatives:onceperturnsave,firstsaveonly,otherCharmdeleted,ordersafterend.',[488]),
},
'Dragon’s Breath': {
 'grant':('PDF126 level2 BONUSAction/TOUCH/VSMhotpepper/Concentration1min;ONEWILLINGcreature,chooseAcid/Cold/Fire/Lightning/Poisonforcasting. Positive:allyTouchchoosePoison; negatives:Thunder,unwilling,remote,freegrant,casterclassentitlementborrowedbyrecipient. Persistownedgranteventandtype.',[]),
 'breath-action':('PDF126 activeacceptedrecipientMAGICACTION15ftConeEACHcreatureDEXsave3d6chosentypefail/halfsuccess;+1d6/slotabove2. Positive:slot3=4d6andrecipientMagicAction; negatives:AttackAction,freebreath,changeeachbreathType,spareallies,aftergrantends,sourceSummonDragonBreathsubstitution.',[]),
},
'Earthquake': {
 'area-events':('PDF127 level8 Action/500ft/VSMfracturedrock/Concentration1min;VISIBLEGROUNDpoint100ftradiusCIRCLEgroundDifficultTerrain. CASTandENDYOURturneachGROUNDEDcreatureareaDEXsavefailProne488ANDConcentrationBROKEN353. Positive:failedsaveterminatesconcentrationwithoutdamagecheck; negatives:airborneaffected,CONsave,startturntrigger,successProne,terrainpersistafterending.',[488,353]),
 'fissures':('PDF127 OPTIONAL1d6fissurestotalENDCASTturn,chosenlocationsNOTunderstructures;EACHdepthD10*10ft,width10ft,edge-to-edgearea. CreatureinspaceDEXsavefailFALL,successmoveswithopeningedge. Positive:D6=2twofissures; negatives:openonlaterturnagain,underbuilding,10ftdepthfixed,successstillfalls,unqualifiedfall/geometryinference.',[]),
 'structures':('PDF127 OPTIONALground-contactstructuresarea50BludgeoningCASTandENDYOURturn;0HPcollapses. Creaturesdistance<=HALFheightDEXsaveFAIL12d6Bludgeoning+Prone488+BURIED,escapeActionDC20STRAthletics;SUCCESSHALFdamageONLY. Positive:40fthighbuildingcreature20ftaway; negatives:21ft,wholeHPguess,STRsave,successburied,freeescape,untrustedstructureHP/height.',[488]),
},
'Enlarge/Reduce': {
 'enlarge':('PDF127 level2 Action/30ft/VSMironpowder/Concentration1min;VISIBLEcreatureORobjectNOTworn/carried. UnwillingcreatureCONsavesuccessNOeffect. EnlargeONEsizecategory,AdvantageSTRchecks/SAVES,itsENLARGEDweaponsORUnarmedStrike779hit+1d4. Positive:Medium->Large,STRsaveAdvantage; negatives:twoCategories,abilityscoreincreased,sameDamageonmiss,spellattack+1d4,freeunwillingnosave.',[779]),
 'reduce':('PDF127 sharedcastpredicate;ReduceONEsizecategory,DISadvantageSTRchecks/SAVES,REDUCEDweaponORUnarmedStrike779hit-1d4minimumONEdamage. Positive:3damageD4=4->1; negatives:0damage,reducebyabilitymod,spellDamagepenalty,advantagechecks,forcewillingtargetsave.',[779]),
 'gear-ending':('PDF127 targetedCREATUREworn/carriedgearRESIZES;DROPitemnormalIMMEDIATE;THROWNweapon/ammunitionnormalAFTERhitORmiss. EndingremovesONLYownedsize/statcontributions. Positive:thrownenlargeddaggerhitusesresizedthenreturns; negatives:unwornobjecttargetshelditem,dropremainslarge,throwshrinksBEFOREhit,permanentinventoryweightrewrite.',[]),
},
'Etherealness': {
 'plane-transition':('PDF128 level7 Action/self/VS/upTo8hNOConcentration;BorderEtherealoverlappingCURRENTplane. CastingalreadyEtherealORnonborderingplaneENDSIMMEDIATELY. Slot8uptoTHREEwillingincludingYOU,9uptoSIX,within10ftcast. Positive:slot8three; negatives:slot8four,unwilling,11ft,planeunknown,bordereligibilityinferred,Concentrationrequirement.',[]),
 'movement-perception':('PDF128 anydirection;UP/DOWNcostEXTRA1ftperfoot;oldplaneseenGRAYmax60ft. Affect/affectedONLYEtherealcreatures/objects/effects;otherscannotperceive/interactexceptACTUALspecialability. Positive:up10cost20; negatives:normalupcost,61ftsight,damageMaterialcreature,automaticTruesightonallobservers,privateomniscience.',[]),
 'return-ending':('PDF128 spellendsreturnCORRESPONDINGoldplanespace;ifOCCUPIEDnearestunoccupiedSHUNTandFORCEdamageTWICEfeetmoved. Positive:15ftshunt30Force; negatives:free10ftchoiceBlinkrule,damage1perfoot,occupiedarrival,leftEtherealpermanent,positionguessed. Persistactualplane/space/ending/displacementdicefreeoutput.',[]),
},
'Eyebite': {
 'repeat-target-ending':('PDF129 level6 Action/self/VS/Concentration1min;ONEVISIBLEcreaturewithin60ftWISsavefailchosenAsleep/Panicked/Sickened. EACHYOURturnMagicActiontargetanother,NEVERtargetagaincreaturethatSUCCEEDEDagainstTHIScast. Positive:priorfailedtargetcanbechosenwherelegal; negatives:priorSUCCESSretarget,freeeveryturn,61ft,saveCON,othercastimmunityinvented.',[]),
 'asleep':('PDF129 acceptedfailedWIS EyebiteAsleep->Unconscious487UNTILdamageANY ORanothercreatureACTIONshakeawake ORspellends. Positive:1damagewakes; negatives:casteronlydamage,freewake,noendingonsourcedamage,clearingunrelatedUnconscious. Exact487endingAPI.',[487]),
 'panicked':('PDF129 acceptedEyebitePanicked->Frightened488;EACHturnDash780awaybysafestANDshortestroute. MOVESspace>=60ftawayANDcannotSEECASTER->effectENDS. Positive:60ftandoccludedendimmediate; negatives:59ft,60ftwithLOS,endturnonly,unsafeGMguessedroute,unrelatedFearended.',[488,780]),
 'sickened':('PDF129 acceptedEyebiteSickened->Poisoned727forSPELLduration. Positive:failedsavePoisoned; negatives:PoisonDAMAGE,additionalDCsaveeachturninvented,poisonimmunityignored,endunrelatedPoisoned. Endthrough504trustedsourceevents.',[727,504]),
},
'Fire Shield': {
 'choice-resistance':('PDF132 level4 Action/self/VSMphosphorusORfirefly/10minNOConcentration;Bright10+Dim10further;WARMshieldColdResistance ORCHILLshieldFireResistance. Positive:warmCold10->5; negatives:warmFireResistance,both,Concentrationrequired,stacksdoublehalve,source106Incapacitateddismissal.',[487]),
 'reactive-damage':('PDF132 wheneverCREATUREwithin5ftHITSyouMELEEattackroll,attacker2d8FIREforWARM OR2d8COLDforCHILL. NOReactioncost. Positive:melee5fthitwarmFire; negatives:miss,rangedattack,reach10fthitfrom10ft,Reactioncharged,allattackersareaDamage,wrongchosenType.',[]),
},
'Flame Blade': {
 'manifest':('PDF132-133 level2 BonusAction/self/VSMsumacleaf/Concentration10min;FREEHANDblade,Bright10+Dim10further;LETGOdisappearsbutmayevokeAGAINBonusActionwhileSPELLongoing. Positive:drop/reinvokewithoutnewslot; negatives:heldfullhands,inventedweaponproficiency,dropendsspell,freeinvoke,normalmetalitemcreated.',[]),
 'attack':('PDF133 ownbladeMAGICACTIONMELEE spellattackhit3d6+castingabilityMODFIRE;+1d6/slotabove2. Positive:slot3CHA3=4d6+3; negatives:AttackActionExtraAttack,STRinsteadCHA,oldnoMODvariant,weaponmasterygrant,hitwithoutblade.',[]),
},
'Flaming Sphere': {
 'placement':('PDF132 level2 Action/60ft/VSMwax/Concentration1min;5ftDIAMETERsphereUNOCCUPIEDGROUNDspace;Bright20+Dim20further;touchunworn/uncarriedflammableobjectignites. Positive:groundplacement; negatives:5ftradius,airborne,occupied,heldobjectignition,lightasdamagewithoutactualevent.',[]),
 'movement-damage':('PDF132 BonusActionrollsphereupto30ftGROUND,barrier<=5fthigh,pit<=10ft;MOVEINTOcreaturespacesaveandSTOPSforthisturn. ENDEACHcreatureturnwithin5ftDEXsave2d6Firefail/halfsuccess;+1d6/slotabove2. Positive:ramthenendturnseparateevents; negatives:oldonceperturncapinvented,flying30ft,6ftbarrier,11ftpit,keepmovingafterram,nearstartturntick.',[]),
 'ending':('PDF132 sphere/light/ownedignitionongoingcontributionseparatefromordinaryfireconsequences;spellendsConcentration/time1min/source106dismissal. Positive:endingremovesSphere/lightwithoutundoingactualobjectburning; negatives:permanentsphere,extinguishunrelatedrealFire,foreigncasterdismissal,Incapacitateddismissal.',[487]),
},
'Flesh to Stone': {
 'restraint':('PDF133 level6 Action/60ft/VSMcockatricefeather/Concentration1min;VISIBLEcreatureCONsave;FAILRestrained727;SUCCESS Speed0UNTILSTARTYOURNEXTturn. CONSTRUCTAUTOsucceeds. Positive:ConstructSpeed0withoutRestrained; negatives:successNOeffect,constructfails,DEXsave,permanentSpeed0,unseen.',[727]),
 'repeat-saves':('PDF133 RestrainedtargetENDitsEACHturnCONsave;trackSUCCESS/FAILURENONconsecutively,THREEsuccessENDSSPELL,THREEfailPETRIFIESforduration. Positive:fail/success/fail/success/failpetrifies; negatives:consecutiverequirement,initialsavecounterunrecorded,3successPetrified,onceperround,multipleadvancesoneevent.',[727]),
 'petrification-ending':('PDF133 own3failurePetrifiedrequiresACTUALsource186admissionowner;maintainConcentrationENTIREpossible1min->PetrifiedUNTILGreaterRestoration624ORsimilarqualifiedmagic. OtherwiseConcentrationendingremovestemporaryPetrified. Positive:fullminutepermanence; negatives:permanentat3failswithoutfullminute,allPetrifiedsources624cleared,PoisonDAMAGEimmunityinventedbycondition,restoreconcentrationafterend.',[624]),
},
'Fly': {
 'speed':('PDF133-134 level3 Action/Touch/VSMfeather/Concentration10min;ONEWILLINGcreatureFlySpeed60ANDHOVER;slotabove3+ONEtarget. Positive:slot4twohover; negatives:unwilling,Speed=walkingSpeed,hoveromitted,slot3two,unsupportedgrantfromtotalclasslevel.',[]),
 'ending-fall':('PDF134 spellendtargetFALLSifALOFTunlessCANSTOPfall. Positive:qualifiedFeatherFallreactionstopsdamage; negatives:autogentlyfloatLevitatevariant,fallwhengrounded,ignoreactualotherFlysource,deleteunrelatedFly,unknownheightinferred. Sourceendingsandfallresolutionremainatomic.',[]),
},
'Freezing Sphere': {
 'burst':('PDF134 level6 Action/300ft/VSMminiaturecrystalsphere/instant;60ftradiusSphereCONsave10d6Coldfail/halfsuccess;+1d6/slotabove6. Positive:slot7=11d6; negatives:DEXsave,30ftradius,301ft,successzero,objectdamagenotstated.',[]),
 'water':('PDF134 strikeBODYwaterfreezes6inchesdeep30ftSQUAREfor1minute. SURFACE-swimmingcreaturestrappedRestrained727;ACTIONSTRAthleticsvsYOURDCescape. Positive:surfaceswimmeractionescape; negatives:underwateralltrapped,6ftdepth,30ftradius,freeescape,normaliceafterminuteunknown. Actualwatergeometryrequired.',[727]),
 'delayed-globe':('PDF134 mayREFRAINfiringoncompletedcast->coolSLINGBULLETsizeglobeinhand;YOUorrecipientmayTHROW40ftORSLINGnormalrange;IMPACTshattersSAMEqualifiedburst/water,maySETDOWNsafely;unshatteredEXPLODESafter1minute. Positive:hand-offthrow40; negatives:41ftthrowwithoutSling,anothercastslotonthrow,settingdowndetonates,permanentstorage,missedminuteexpiry.',[]),
},
'Gaseous Form': {
 'transform':('PDF134 level3 Action/Touch/VSMgauze/Concentration1h;WILLINGcreatureANDworn/carriedgearMIST,slotabove3+ONEtarget. B/P/SResistance,ProneIMMUNITY,STR/DEX/CONsaveAdvantage. Positive:magicalBludgeoningalsoResisted; negatives:allDamageResistance,Proneadmission,unwilling,gearleftusable,oldnonmagic-onlyvariant.',[488]),
 'movement-restrictions':('PDF134 ONLYmovementFly10ftHOVER;canenter/occupyOTHERcreaturespace,passNARROWopenings,LIQUIDSsolid. CannotTALK/manipulateobjects/drop/usecarriedgear/ATTACK/CAST. Positive:1inchopenpathwithtrustedgeometry; negatives:swimthroughwater,dooropenbyhands,castSubtlewhilemist,ordinary30ftwalk,hoverremoved.',[]),
 'ending':('PDF134 spellendsONTARGETat0HP ORtargetMAGICACTIONendself;normalConcentration/timeendingalso. Source106CASTERownspellnoactiondismissalifNOTIncapacitatedseparatefromTARGETMagicAction. Positive:targetselfMagicAction; negatives:targetnoactionselfendwhenothercaster,0HPspellpersists,undoHPdamage,clearotherForm,Incapacitatedcasterdismissal.',[487]),
},
'Gate': {
 'portal':('PDF135 level9 Action/60ft/VSMdiamond5000+GPnotconsumed/Concentration1min;VISIBLEUNOCCUPIEDspaceportalDIFFERENTplanePRECISElocation,circular5-20ftdiameterANYorientation;destinationVISIBLEthroughportal. FRONTonlytravel->nearestUNoccupiedotherside. Planarruler/deitymayPREVENTpresencedomainopening. Positive:frontcrossingknownplane; negatives:sameplane,backcross,21ftdiameter,4999GP,ignoreexplicitrulerblock,unverifieddestination.',[]),
 'named-creature':('PDF135 WHENCASTspeakSPECIFICcreatureTRUENAME,notpseudonym/title/nickname;mustbeOTHERplane->portalnexttoit,transportsnearestUNOCCUPIEDcasterside. NOcontrol/compulsionafterarrival;GMactualbehavior. Positive:authorizedtrueidentityotherplane; negatives:sameplane,nickname,autoobedience,secrettrueNameguess,hostileAIactionsinferred. Unknownplanaridentityunavailablewithoutleak.',[]),
},
'Globe of Invulnerability': {
 'barrier':('PDF136-137 level6 Action/self/VSMglassbead/Concentration1min;IMMOBILE10ftEmanationinitialcasterlocation. SpellsCASTOUTSIDElevel<=5cannotAFFECTanythingWITHIN;targetingstillLEGALbutNOeffect,areasEXCLUDEinside. Positive:outsideFireballinsideprotected; negatives:movebarrierwithcaster,blocksourcecastINSIDE,blankettargetadmissionban,blocknonspellattack.',[]),
 'level-interaction':('PDF137 slotabove6+ONEblockedspelllevel:slot7blocks<=6. ActualCASTEFFECTIVEspelllevelmustselectlevel,notbaselevel/name; targetspellslotconfirmednoeffectstillspendswhereapplicable. Positive:slot7blockslevel6; negatives:slot6blocks6,upcast7spelltreatedbase1,spellsourceinsideblocked,inferprivatecastlevelinpreview.',[]),
},
'Grease': {
 'area':('PDF137 level1 Action/60ft/VSMporkrindORbutter/1minNOConcentration;10ftSQUAREGROUNDnonflammableGREASE,DifficultTerrain. Positive:trustedgroundsquare; negatives:Sphere,flammablegreaseignition,airbornearea,Concentrationrequired,permanentterrain.',[]),
 'prone-ending':('PDF137 APPEARANCEeachstandingcreatureDEXsavefailProne488;ALSOcreatureENTERSorENDSturthereDEXsavefailProne. Positive:entrythenendturnbothsaves; negatives:onceperturncapinvented,startturninsteadend,airborneProne,saveCON,eraseProneautomaticallywhenleaveGreasewithoutsource. TerrainenddoesnotundoalreadyadmittedProne.',[488]),
},
'Haste': {
 'benefits':('PDF139 level3 Action/30ft/VSMlicoriceroot/Concentration1min;VISIBLEWILLINGcreatureSpeedDOUBLE,AC+2,DEXsaveAdvantage,extraActionEACHitsTURN. Positive:Speed30->60AC15->17; negatives:unwilling,unseen,doubleAC,allSavesAdvantage,stackduplicateHaste. Sourceincapacitationrulesretainownauthority.',[]),
 'limited-action':('PDF139 ONLYextraHasteAction:AttackONEattack ORDash780 ORDisengage781 ORHide1137 ORUtilizequalifiedconcreteoutput1139SignalWhistle. OtherunknownUtilizeeffectsremainunavailableuntilnamedowner;doNOTcall1139genericUtilize. Positive:oneweaponattackdespiteExtraAttack,qualifiedWhistle; negatives:MagicAction,multiattack,secondExtraAttack,freeBonusAction,unknownitemeffect,classRoguegrantimported.',[780,781,1137,1139]),
 'ending-lethargy':('PDF139 whenHasteENDS,targetIncapacitated487+Speed0 UNTILENDitsNEXTturn. Positive:Concentrationbreakimmediatelystartslethargy; negatives:onlynormalexpiry,skipnextturnending,endSTARTturn,automaticallactoractionsbeforeending,unrelatedIncapacitatedcleared. Noinventedoptionalactivation.',[487]),
},
})

SPELLS.update({
'Hold Monster': {
 'paralysis':('PDF140 level5 Action/90ft/VSMstraightiron/Concentration1min;ONEVISIBLEcreatureWISsavefailParalyzed727;+ONEtarget/slotabove5. Positive:Undeadcreaturecanqualify; negatives:oldUndeadimmunityimported,Humanoidonly,CONsave,91ft,successfulsaveParalyzed,wrongtargetcount.',[727]),
 'repeat-save-ending':('PDF140 ENDtargetEACHturnrepeatWISsave,successENDSspellONLYitself;Concentration/timeendsownedinstances. Positive:slot6tworecipientsoneends; negatives:casterturnsave,allParalyzedcleared,actioncost,endafteranydamageinvented. Consume504trustedParalyzedending.',[504]),
},
'Hypnotic Pattern': {
 'admission':('PDF141 level3 Action/120ft/SMconfetti/Concentration1min;30ftCubeMOMENTARYpattern,ONLYcreaturesAREAthatcanSEEpatternWISsavefailCharmed488 ANDwhileCharmedIncapacitated487+Speed0. Positive:failedsaveknownobserver; negatives:Vrequirement,Blindedviewer,observeroutsideCube,failedsaveonlyCharmwithoutriders,inferhiddensight.',[488,487]),
 'ending':('PDF141 targetANYdamageORsomeoneELSEACTIONshakeENDSownrecipient;Concentration/timealsoend. Positive:1damagefromenemyends; negatives:onlycasterdamage,endatstartturnwithouttrigger,freeShake,selfshakewhileIncapacitated,clearotherCharm/Incapacitatedsource.',[488,487]),
},
'Incendiary Cloud': {
 'area-movement':('PDF142 level8 Action/150ft/VS/Concentration1min;20ftradiusSphereHeavilyObscuredembers/smoke;STRONGwinddisperses;STARTYOURturnmoves10ftAWAYfromyouINDIRECTIONYOUCHOOSE. Positive:choiceamongawayvectors; negatives:towardcaster,10minCloudkillduration,freeextraMove,ordinarybreezeending,unverifiedwind/position.',[1061]),
 'damage':('PDF142 APPEARANCEEACHcreatureDEXsave10d8Firefail/halfsuccess;alsoSphereMOVESINTOspace,creatureENTERSorENDSturthere,ONLYONCEperTURNpercreature. Positive:moveandentrysameturnoneevent; negatives:CONsaveCloudkillvariant,onceROUND,11d8upcastinvented,standingstartturntick,missinginitialdamage.',[]),
},
'Invisibility': {
 'admission':('PDF143 level2 Action/Touch/VSMeyelashgum/Concentration1h;ONEcreatureInvisible962;+ONEtarget/slotabove2. Positive:slot3tworecipients; negatives:inventedwillingrequirement,remoteTouch,allobserversblind,grantHideendingsemantics,makingallgearunownedfactsvisible. Exactobserver962API.',[962]),
 'ending':('PDF143 ENDSearlyIMMEDIATELYAFTERtargetmakesATTACKROLL ORDEALSDAMAGE ORCASTSSPELL. Positive:missattackendsAFTERroll; negatives:onlyhit,onlyhostilespell,endingBEFOREattackbenefit,onlyturnboundary,GreaterInvisibilitysametriggers,removeotherInvisible. Exactperrecipient/sourceendings.',[962]),
},
'Mage Hand': {
 'manifest':('PDF145 cantrip Action/30ft/VS/1minNOConcentration;floatinghandchosenpoint,VANISHESifEVER>30ftFROMYOU ORRECAST. Positive:castermovesawayhandvanishes; negatives:30ftonlyinitialcheck,Concentration,secondhandkeepsfirst,permanenthand,casttargetunknownspace.',[]),
 'manipulation':('PDF145 CASTandlaterMAGICACTIONmanipulateOBJECT/openUNLOCKEDdoor/container/stowretrieveOPENcontainer/pourVIAL;asPARTactionmovehandupto30ft. NOattack/MAGICITEMactivation/carry>10lb. Positive:openunlockeddoorandmove20; negatives:lockpickingwithoutownsource,11lb,attack,magicWandactivation,BonusActioncontrol. Actualobjectownership/worldfactsrequired,noNPCinference.',[]),
},
'Major Image': {
 'image':('PDF147-148 level3 Action/120ft/VSMfleece/Concentration10min;VISIBLEspot,image<=20ftCube,visual+sound+smell+temperature;CANNOTdealDAMAGEorCAUSECONDITIONS. Positive:nonharmfulheat/sound; negatives:21ftCube,illusionFireactualdamage,Charmedbyimage,unseeninitialspot,actualcreaturecreated.',[]),
 'movement-interaction':('PDF148 IFYOUwithinspellrangeMagicActionmoveimagewithinrange,alterappearance/soundsnaturallyinclconversation. Physicalinteractionreveals;StudyINTInvestigationvsDCdiscerns;discernerviewseesthrough/othersensesFAINT. Positive:oneobserverdiscernsothersnot; negatives:freeMove,outofrangecontrol,allobserverslearnprivatecheck,soundpermanentlyreal.',[]),
 'upcast':('PDF148 level4+slot lastsUNTILDISPELLEDWITHOUTConcentration. Positive:slot4persistentimage; negatives:oldslot6minimum,slot3permanent,stillConcentrationat4,broaden20ftCubeatupcast,inventedpermanentdamage. Keepmovement/interactionrestraintsandqualifiedDispel355/965/966API.',[355,965,966]),
},
'Mass Suggestion': {
 'suggestion':('PDF148 level6 Action/60ft/VMsnaketongue/24hNOConcentration;<=TWELVEVISIBLEcreaturesHEAR/UNDERSTAND SAMEcourse<=25wordsACHIEVABLEandnotobviousdamageANYtarget/ALLIES. EACHWISsavefailCharmed488pursuebestability. Positive:12independentsaves; negatives:13,unheard,26words,allyharm,Concentrationinvented,GMmodelcommandsinferred.',[488]),
 'duration-ending':('PDF148 ownrecipientENDSonactivityCOMPLETE ORcaster/ALLIESdamageit ORduration24h;slot7TENdays,8THIRTYdays,9THREEHUNDREDSIXTYSIXdays. Positive:slot9366days; negatives:365,damageotherenemyends,allrecipientsonetargetcompletion,lastingaftercompletedactivity,foreignCharmending.',[488]),
},
'Minor Illusion': {
 'sound':('PDF149 cantrip Action/30ft/SMfleece/1min;chooseSOUND ORimage,RECASTendsprior. Soundwhisper->scream,chosenvoice/roar/drums,continuousORdiscretebeforeend. StudyINTInvestigationvsDCdiscerns->FAINTforobserver. Positive:intermittentbell; negatives:Vcomponent,soundANDimageonespell,Recastkeepsprior,physicalsounddamage,mindreadingvoiceunknown.',[]),
 'image-interaction':('PDF149 IMAGEofOBJECT<=5ftCube,NOsound/LIGHT/smell/othersense;physicalinteractionrevealsnotSolid. StudyINTInvestigationvsDCdiscerns->FAINTforobserver. Positive:5ftchestimage; negatives:creatureanimationfromobjectscope,6ftCube,lightillumination,blockingprojectile,freeillusiondisclosuretoall. Same1min/recastending.',[]),
},
'Mirror Image': {
 'images':('PDF150 level2 Action/self/VS/1minNOConcentration;THREEillusoryduplicatesmove/mimicactions;ignoreALLotherdamage/effects. All3destroyedENDSspell. Blind960/Blindsight/TruesightattackerUNAFFECTED. Positive:blastdoesnotdestroyimages; negatives:separatecreaturesHP,onlyonespawned,oldAC10+DEXimage,unqualifiedspecialsensebypass,inventConcentration.',[960]),
 'attack-interception':('PDF150 EACHtimecreatureHITSYOUattackrollrollD6PERREMAININGduplicate;ANY>=3ONEduplicatehitINSTEADanddestroyed. Positive:3imagesroll1/3/6onlyONEdestroyed,realactoruntouched; negatives:D20oldvariant,rollonmiss,destroyallqualifyingimages,2countsasredirect,privateoutcomeknowledgeassumed,specialsenseattackerredirected.',[]),
},
'Move Earth': {
 'area-shaping':('PDF151 level6 Action/120ft/VSMminiatureshovel/Concentration2h;area<=40ftSIDE,DIRT/SAND/CLAY;shapeelevation/trench/wall/pillar<=HALFareaLARGESTdimension. Changescomplete10MINUTES;each10minConcentrationmaychooseneWareawithinrange. Positive:40ftarea20ftraise; negatives:21ftraise,instanttrap,neWareaeveryturn,unknownterraininference.',[]),
 'restrictions':('PDF151 NOTnaturalSTONE/stoneconstruction,plantscarriedNOTgrowthmodified;rocks/structuresSHIFTtoaccommodate,unstablestructureMIGHTcollapseonlytrustedGMresolvedfact. Positive:dirtwithplantmovesplant; negatives:reshapegranite,instantcreaturedamage,trappedwithoutactualqualifiedfact,automaticstructurecollapse,inventgrowthyield. Persistauthorizedterrain/worlddeltasandrollback.',[]),
},
'Plane Shift': {
 'willing-transport':('PDF153 level7 Action/Touch/VSMforkedmetalrod250+GPATTUNEDDESTINATIONplane/instant;YOU+uptoEIGHTWILLINGcreaturesLINKHANDSINCIRCLE->DIFFERENTplane;generalcity/palacedestinationGMdeterminesINORNEAR. ORKNOWNsigilsequenceotherplanecircle;overflowclosestunoccupiedadjacentcircle. Positive:9totalincludingcaster; negatives:10total,unwillingattackbanishmentOLDVARIANT,missinglinkedhands,unattunedrod,250GPtotalgearwithoutrod,arbitraryexactarrivalwithoutGM.',[]),
},
'Power Word Stun': {
 'admission':('PDF154 level8 Action/60ft/V/instant;ONEVISIBLEcreatureauthoritativeHP<=150->Stunned777;HP>150Speed0UNTILSTARTYOURNEXTturn. NOinitialsave. Positive:150Stunned,151Speed0; negatives:151noeffect,initialCONsave,publicprivateHPdisclosure,HPclientforgery,Concentrationinvented.',[777]),
 'repeat-save-ending':('PDF154 StunnedtargetENDitsEACHturnCONsave,successendsONLYthisStunnedinstance504. Positive:firstendturnsuccess; negatives:startturn,repeatWIS,infiniteStunnedwithoutsave,clearallStunned,Speed0fallbackpermanent. FallbackendingisdistinctfromStunnedconditionending.',[504]),
},
'Prestidigitation': {
 'instant-sensory-clean':('PDF154 cantrip Action/10ft/VS/upTo1h;instantHARMLESSsensorysparks/wind/music/odor ORlight/snuffCANDLE/TORCH/SMALLcampfire ORclean/soilOBJECT<=1CUBICft. Positive:snuffsmallcampfire; negatives:damagingfire,largebonfire,2cubicftclean,creaturecleanedunderobjectpredicate,permanentmusic. Sourceoptionsonepercast.',[]),
 'lasting-effect':('PDF154 MinorSensationchill/warm/flavor<=1cubicftNONLIVINGmaterial1h ORMagicMarkcolor/smallmark/symbolobject/surface1h. AcrossrecastsMAXTHREEnoninstantaneouseffects. Positive:thirdmarkallowed; negatives:fourthactive,livingskinchilledasmaterial,2cubicft,permanentmark,recastautomaticallyendallinsteadcap. Recordexactcastereffectownership.',[]),
 'trinket-image':('PDF154 MinorCreationnonmagicTRINKET ORILLUSORYIMAGEfitsHAND,UNTILENDYOURNEXTturn;trinketNOdamage/NOMONETARYworth. Positive:harmlesshandtrinket; negatives:valuablegold,weaponDamage,Largeobject,1htrinket,realimagephysicalbarrier. Countsnoninstantaneousmax3.',[]),
},
'Prismatic Spray': {
 'rays1-5':('PDF154-155 level7 Action/self/VS/instant;60ftConeEACHcreatureDEXsaveANDD8rayselection;1Fire/2Acid/3Lightning/4Poison/5Cold EACH12d6failure/HALFsuccess. Positive:ray2DEXfail12d6Acid; negatives:casterchoosesray,onlyoneD8foralltargets,10d6oldvariant,ray5Fire,slotupcastinvented.',[]),
 'indigo':('PDF155 D8=6andFAILEDinitialDEX->Restrained727;ENDtargetEACHturnCONsave,3SUCCESSendscondition,3FAIL->PetrifiedUNTILGreaterRestoration624orsimilarqualifiedeffect. NONconsecutivecountersbothsaved. Positive:F/S/F/S/Fpetrifies; negatives:successfulDEXstillrider,consecutiveonly,initialDEXcountedasCONfailure,temporaryPetrified1mininvented,deleteallrestraints.',[727,624]),
 'violet':('PDF155 D8=7andFAILEDinitialDEX->Blinded960;STARTCASTERNEXTturnTARGETWISsave,SUCCESSendsBlinded963,FAILalsoENDSBlindedandteleportsOTHERplaneGMchoice. Positive:failedWISendsblind+trustedplanejump; negatives:failurepermanentBlind,saveENDtargetturn,destinationAIchoice,sameplanetarget,endforeignBlind. Atomicconditionandplaneeffect.',[960,963]),
 'special8':('PDF155 D8=8->TWOrays,ROLLTWICE REROLLANY8. EachsameinitialDEXoutcomeappliessourceoutcomes;rollorder/dicepersist. Positive:8then2/7executesAcid+Violet; negatives:8recursivelydoublingforever,onlyoneray,choosecolors,rerollinitialDEX,silentlytruncatelegal8rerolls. Injectableboundedtestrandomstream,productioninvaliddiceproviderfailswithoutpartialmutation.',[]),
},
'Reverse Gravity': {
 'area-fall':('PDF159 level7 Action/100ft/VSMlodestone+ironfilings/Concentration1min;50ftradius100ftHIGHCylinder,UNANCHOREDobjects/creaturesFALLUPtoTOP;collisionCEILING/anchoredobjectasordinaryfalldamage,otherwiseHOVERtop. Positive:unanchoredfallsup100; negatives:anchoredmoves,100ftradius,automaticnodamageceiling,onlycreatures,worldgeometryinferred.',[]),
 'restrained-object-save':('PDF159 creaturemayDEXsaveGRABREACHABLEFIXEDobjectavoidUPfall. Positive:sourceknownreachableanchor+success; negatives:noanchorreachableautosave,sameasRestrainedcondition,STRsave,anchorcreatedbyinference,forcedsaveforlooseobjects. Noinventedpostgrabcondition.',[]),
 'ending-fall':('PDF159 spellENDaffectedobjects/creaturesFALLDOWN. Positive:tophover->sourcefall; negatives:gentlyfloat,remainhovering,eraseactualotherflight/qualifiedfallprevention,untrustedheight,duplicatefallDamagereceipt. Atomicposition/condition/damagehistory.',[]),
},
'See Invisibility': {'perception':('PDF160 level2 Action/self/VSMtalc/1hNOConcentration;SEEcreatures/objectsWITHInvisible962asVISIBLE andSEEETHEREALghostlycreatures/objects. Positive:visibleauthorizedInvisibleobject; negatives:endInvisiblecondition,grantTruesight/allIllusionbypass,seeTotalCover,seeunboundedhiddenfacts,otherobserversseethem. Senseadmissionobserver-specific.',[962])},
'Seeming': {
 'appearance':('PDF160 level5 Action/30ft/VS/8hNOConcentration;chosenVISIBLEcreatures,UNWILLINGmayCHAsaveSUCCESSunaffected. SameORdifferentillusionbodies/gear,upto1ftshorter/taller/heavier/lighter,SAMEbasiclimbs. Positive:twochosencreaturesdifferentappearance; negatives:WISsave,2ftchange,statmutation,unseen,alltargetsneedforcedsave.',[]),
 'inspection-ending':('PDF160 physicalinspectionFAILSillusionphysicalconsistency;StudyINTInvestigationvsDCsuccessbecomesAWAREDISGUISE. Time/source106owncasterdismissalendsownedspell,notGM-inferredbelief. Positive:touchhatpassesthrough; negatives:hatblocksrealobject,automaticmentalcontrol,allobserversprivatecheckdisclosure,alteredactualrace. Noinventedattackendsillusion.',[]),
},
})

SPELLS.update({
'Slow': {
 'target-admission':('PDF163 level3 Action/120ft/VSMmolasses/Concentration1min;chooseuptoSIXcreaturesin40ftCube,EACHWISsavefailaffected. Positive:sixindependentsaves; negatives:7targets,Sphere,allcreaturesmandatory,CONsave,494weaponmasterySlowalias. Exactsourceidentityseparatefrommastery.',[]),
 'turn-limits':('PDF163 affectedSpeedHALF,AC-2,DEXsave-2,NOReactions. OnOWNturnACTIONORBonusActionNOTboth;AttackActionONLYONEattack. CastingwithSOMATIC25%FAILchance. Positive:Sspellinjectedpercent25fails; negatives:nonSomatic25%failure,oldD20nextturncastdelay,halfattacksinstead1,allowBA+Action,otherSaves-2. Persistpercentdrawandfailedcastcostreceiptwithoutinventedrefund.',[]),
 'repeat-save-ending':('PDF163 ENDaffectedtargetEACHturnWISsave,successendsownrecipient;Concentration/timeendmodifiers. Positive:oneof6endsotherscontinue; negatives:startturnsave,clearweaponSlow494,permanentACpenalty,duplicateending. Recomputeownedmodifierswithoutundoingunrelatedsources.',[]),
},
'Sunbeam': {
 'line-damage-blindness':('PDF166 level6 Action/self/VSMmagnifyingglass/Concentration1min;5ftWIDE60ftLONGLine,EACHcreatureCONsaveFAIL6d8Radiant+Blinded960UNTILSTARTYOURNEXTturn;SUCCESSHALFdamageONLY. Positive:failedsaveblindnextsourceboundary; negatives:DEXsave,successfulsaveBlind,12d6Sunburstimported,endtargetturn,undeadextraDamageinvented.',[960,963]),
 'repeat-action-light-ending':('PDF166 UNTILendsMAGICACTIONcreateNEWLine,samequalifieddamage/blindness;castermoteBright30+Dim30further,SUNLIGHT. Positive:nextturnnewLinewithMagicAction; negatives:BonusActionline,ordinarynotSunlight,freeeachturnline,upcastdamageinvented,sourceendingclearsunrelatedBlind/light. PersistindividualLineeventownership.',[960,963]),
},
'Sunburst': {
 'burst':('PDF167 level8 Action/150ft/VSMsunstone/instant;60ftradiusSphereSUNLIGHT,EACHcreatureCONsavefail12d6Radiant+Blinded960ONEminute,successHALFdamageONLY. Positive:faildamage+Blind; negatives:DEXsave,151ft,Concentration,successfulsaveBlind,oldUndeadsaveDisadvantageimported.',[960]),
 'blindness-repeat-save':('PDF167 ownBlindedcreatureENDitsEACHturnCONsave,successendsONLYsourceBlind963;timeout1minalsoends. Positive:onerecipientendsindependently; negatives:casterturn,allBlindsremoved,repeatWIS,firstsaveonly,sourceinstantaneousmeansBlindforever.',[963]),
 'darkness-ending':('PDF167 DISPELSDARKNESSinAREAcreatedbyANYspellregardlesslevel. Positive:level8magicalDarknessoverlapendedexactscope; negatives:nonspellnightremoved,level3capDispelMagicimported,outofareaDarknesswiped,unrelatedsourceeffectsended. ConsumequalifiedsourceDarknessidentity/areaendingAPI,not353concentrationeventasDispel.',[]),
},
'Telekinesis': {
 'creature':('PDF167 level5 Action/60ft/VS/Concentration10min;CASTandlaterMagicActionONEVISIBLEHugeorsmallerCREATURESTRsavefailmoveupto30ftANYdirectionWITHINspellrange,Restrained727untilENDYOURNEXTturn,suspendediflifted;thenFALLunlessrepeatoptionandFAILsaveagain. Positive:STRfailHuge30ft; negatives:Gargantuan,oldcontestcheck,saveDEX,move31ft,outside60ft,successfulsaveRestrained.',[727]),
 'unheld-object':('PDF167-168 sameactivecast,ONEVISIBLEHugeorsmallerUNWORN/UNCARRIEDobjectAUTOMATICmoveupto30ftANYdirectionwithin60ft. Positive:Hugeunheldobject30ft; negatives:weight1000lbcap2014invented,Gargantuan,heldobjectautosuccess,31ftmovement,unverifiedobjectidentity.',[]),
 'held-object':('PDF168 WORN/CARRIEDHugeorsmallerobject:HOLDERSTRsave,FAILpullawayandmoveupto30ftANYdirectionwithinrange. Positive:failedholderSTRweaponpulled; negatives:contest2014,forcedwillingnessgate,successstillpull,unknownforeigninventorydisclosure,wholecreaturemovedinobjectmode.',[]),
 'fine-control-ending':('PDF167-168 activeONEtargetcanrepeatORswitch;SWITCHpriorNOLONGERaffected. FinecontrolobjectsONLYsourceexamplemanipulateSIMPLEtool;sourceendsmid-sentenceanddoesNOTauthorizeunknownremotecomplexactions. Positive:sourceboundownedtoolmanipulation; negatives:inventedlocktrapspellitemexecution,noActioncontrol,switchretainRestrained,afterendingkeepunsupportedworldmodification. Requireconcretequalifiedtool/actionoutputforanyactualoperation.',[]),
},
'Teleport': {
 'destination-category':('PDF168 level7 Action/10ft/V/instant;YOU+uptoEIGHTWILLINGVISIBLEcreatures ORONEVISIBLELargeorsmallerOBJECTnotheld/carriedbyUNWILLINGcreature. DestinationKNOWN,SAMEplane. Categories:PermanentCircleknownsigils;LinkedObjectTAKENdestinationwithinLASTSIXMONTHS;VeryFamiliaroftenvisited/carefullyStudied/currentlySEE;SeenCasually>ONCE;ViewedONCEorDescribed;FalseNONEXISTENT. Positive:ninecreaturestotal; negatives:10,unseen,unwillingholder,otherplane,oldstaledestinationobjectaccepted,arbitrarycategoryupgrade.',[]),
 'mishap-table':('PDF168-169 GM-authorizedD100:PermanentCircle/LinkedObjectALLOnTarget;VeryFamiliar1-5Mishap/6-13Similar/14-24Off/25-100On;SeenCasually1-33M/34-43S/44-53O/54-100On;ViewedOnceOrDescribed1-43M/44-53S/54-73O/74-100On;False1-50M/51-100S. MishapEACHtravelerORobject3d10Force thenREROLLtable,mayMULTIPLEmishaps. Positive:VeryFamiliar5mishapthen25On; negatives:00=0not100,skipmishapdamage,oneDamageforgroup,truncatelegitimateMishapchain,unsaveddiceonretry.',[]),
 'arrival':('PDF169 OnTargetintendedspot;SimilarCLOSESTvisually/thematicallysimilarACTUALGMplace;OffTarget2d12MILES+D8direction1E/2SE/3S/4SW/5W/6NW/7N/8NE. Positive:D12=4+5andD8=7=>9milesN; negatives:percentdistance2014,rollNfor1,destinationsinventedbyAI,sameplaneconstraintbypassed,unknownnearestspaceinferred. Atomicalltravelerpositions/damage/history,sourceprivatedisclosurefilter.',[]),
},
'Teleportation Circle': {
 'sigil-material-creation':('PDF168 level5 1MINUTE/10ft/VMrareinks50+GP CONSUMED/1round;draw5ftradiusGROUNDcirclelinkKNOWNsigilsPERMANENTcircleSAMEplane. FirstgainabilitylearnTWOdestinationsMATERIALplaneGMdetermined;newsequenceMEMORIZEby1minuteStudy. Positive:consumed50GPandknownsigils; negatives:Srequired,unknownsigilsinferred,otherplane,freeink,temporarydestinationcircle,0minstudy.105longcastingprotocol.',[]),
 'portal-travel':('PDF168 portalopensuntilENDYOURNEXTturn;ANYcreatureENTERSinstantlywithin5ftdestinationcircle ORnearestUNoccupiedspaceifthatspaceoccupied. Positive:noncastercreatureentrybeforeexpiry; negatives:mustbecaster,willingnessgateinvented,afterexpiry,GMunknownnearestspace,oldopencircle24h. Preserveclock/collision/planeauthority.',[]),
 'permanence':('PDF168 SAMElocationcastEVERYDAYfor365days->PERMANENTcircle. Positive:365distinctsourcequalifieddailycastreceipts; negatives:364,365sameDay,locationdrift,missingday,silentcalendarinference,unpaidmaterials. Durableprogress/reload/retry/concurrencywithactualsourcecalendar;unqualifiedworldtimeunavailable.',[]),
},
'Time Stop': {
 'extra-turns':('PDF169 level9 Action/self/V/instant;D4+1consecutiveOWNturns,NOtimepassesOTHERcreatures;normalactions/movementqualifiedsourceonly. Positive:D4=2threeownturns; negatives:D4only,othercreaturetimersadvanceeachturn,freeactioneconomy,unsaveddicereroll,productionhighlevelgrant.',[]),
 'ending-triggers':('PDF169 endsifANYownAction ORcreatedEFFECTaffectsOTHERCREATURE ORobjectWORN/CARRIEDbySOMEONEELSE;alsoMOVE>1000ftfromCASTlocation. Positive:1000ftcontinues1001ends; negatives:onlydirectAttackends,delayedcreatedeffectignored,otherheldobjectallowed,selfobjecttriggers,remainingextraturnsafterending. Exacttriggertransactionendsbeforeanyillegalfurthertime-frozenturn.',[]),
},
'True Seeing': {'senses-disclosure':('PDF171 level6 Action/Touch/VSMmushroompowder25+GP CONSUMED/1h;WILLINGcreatureTruesight120ft. Source190TruesightpermitsseeInvisible962,normalANDmagicalDarkness,VISUALillusionsTRANSPARENTandAUTOsucceedSAVESagainstthem,TRUEformofMAGICtransformedCREATUREorOBJECT,andEtherealwithinrange;NOopaqueTotalCoveromniscience. Positive:sourcequalifiedInvisibleat120; negatives:121ft,unwilling,24GP,unconsumedpowder,otherobserversreceivefacts. Actualsense/cloak/worldsourcepermissionsonly.',[962])},
'True Strike': {
 'weapon-attack':('PDF171 cantrip Action/self/SMweaponPROFICIENTandworth>=1CP;ONEattackwithTHATweapon,castingabilityinsteadSTR/DEXforATTACKandDAMAGE;chooseRADIANTorNORMALweaponDamage. Positive:CHAweaponrangedattackwherequalified; negatives:Vrequirement,unproficient,0CPconjuredtrinket,differentweapon,extraAttackAction,oldAdvantage-nextturnvariant. Nativeweaponattackoutputqualifiedseparately.',[]),
 'damage-scaling':('PDF171 TOTALlevels5/11/17+1/2/3d6EXTRARADIANTwhetherbaseRadiantORnormaltype. Positive:total5normalSlashingPLUS1d6Radiant; negatives:onlyRadiantbasegetsbonus,changedallbonustoSlashing,sourceclasslevelscaling,extraDamageonmiss,level4bonus. Separatemitigationpertypeandonehitreceipt.',[]),
},
'Vampiric Touch': {
 'attack-healing':('PDF171 level3 Action/self/VS/Concentration1min;MELEEspellattackONEcreaturewithinreach,hit3d6NecroticandhealSELF HALFNECROTICdamageACTUALLYdealt(rounddown). +1d6/slotabove3. Positive:7rolledagainstResistance3dealt->1heal; negatives:healhalfrolled7->3despiteresistance,healally,damageIMMUNEtargetbutheal,missheal,range30oldvariant.',[]),
 'repeat-action-ending':('PDF171 whileSPELLongoingEACHYOURturnMAGICACTIONattackagainSAMEordifferentcreature,nonewslot. Positive:turn2MagicActiononeattack; negatives:BonusAction,ExtraAttackActionattacktwice,afterConcentrationend,newslotcostperattack,casterconcentrationrequirementforordinaryfeatselection. Exactownedtime-span106dismissalifnotIncapacitated.',[487]),
},
'Wall of Fire': {
 'placement-damage':('PDF172 level4 Action/120ft/VSMcharcoal/Concentration1min;SOLIDsufaceOPAQUEwall<=60long20high1thick ORRING<=20DIAMETER20high1thick. APPEARANCEeachcreatureAREADEXsave5d8Firefail/halfsuccess;+1d8/slotabove4. Positive:ring20diameter; negatives:20radius,floatingwall,61long,clearwall,missinitialsave,allcreatureswithin10fthaveinitialdamage.',[]),
 'hot-side':('PDF172 chooseONEsideWHENCASTING;ENDcreatureturnwithin10ftTHATsideORINSIDEwall5d8FireNOSAVE,+1d8/slotabove4;OTHERsideNOdamage. Positive:hot-side10ftendturnhit; negatives:11ft,swapfreeeachturn,bothsidestick,savehalfhot-side,anycreaturestartsTurntick. Actualgeometrysidepersisted.',[]),
 'movement-ending':('PDF172 ENTERWALLFIRSTtimeonTURN OR ENDturnINSIDE dealsSAMEqualifieddamageNOSAVE;notgenericonce-per-turncapthatwoulddropendturnevent. Positive:entryandendinsidequalifiedevents; negatives:every1ftmultipleentrydamage,safe-sideoutsideareaentrytick,expiryremovesunrelatedFire,foreigncasterdismissal. NormalConcentration/time1min/source106casterdismissalboundaries.',[487]),
},
})

# Shared execution primitives missing from prior whole contracts; these own only
# the named bounded source behavior, not generic world/condition qualification.
SEAMS = {
 'condition-petrified': ('PDF186 Petrified: creature+NONmagicworn/carriedgearINANIMATE,weightTENtimes,NOaging;Incapacitated487,Speed0CANNOTincrease,attacksAGAINSTAdvantage,AUTOfailSTR/DEXsaves,ALLdamageResistance,PoisonedIMMUNITY. Positive:Fire10->5andPoisoned727rejected; negatives:PoisonDAMAGEimmunityinvented,magicitemtransformed,otherSavesautofail,doubleResistancehalving. Sourceending624removesONLYexacttrustedPetrifiedinstance,restoreownedform/weightwithoutundoingHPdamage.', [487,727,624]),
 'zombie-stat-actions': ('PDF343-344 exactZombie178:MediumUndeadNeutralEvil,AC8,HP15(2d8+6),Speed20,Initiative-2,STR13/DEX6/CON16/INT3/WIS6/CHA5,WISsave+0;PoisonDamage/Exhaustion/PoisonedIMMUNE,Darkvision60,Passive8,understandsCommon+ONEotherlanguage,CANNOTspeak,CR1/4PB2. SlamMELEE+3reach5,hit1d8+1Bludgeoning. Positive:D8=4->5; negatives:OgreZombie85HP,Wyvernattack,CHAspellattackmodifier,Poisonedapplied,verbalcommandsfromZombie. Actualstat/actionqualifiedbeforeFingerDeathcreation.', [178,727]),
 'zombie-undead-fortitude': ('PDF343 ownexactZombie:damagewouldreduce0HP->CONsaveDC5+DAMAGEtaken,successHP1INSTEAD,EXCEPTRADIANTdamageORCRITICALhitNOsave. Positive:damage7DC12CONsuccess1HP; negatives:Radiantstillroll,criticalstillroll,DC5only,multipledeathretries,HP5. SaveownedprovablysameDamageevent,atomicdeath/spawn/HP/history;doesnotgeneralizeallUndeadstatblocks.', [178]),
}

# These four bounded runtime adapters extend shared receipt APIs; 273/277 are
# planning prerequisites only and never supply their executable behavior.
SOURCE_PROTOCOLS = {
 'cast-components': ('PDF104-106/206/244 finite245identityboundary: validate normal classCAST sourceadapter suppliesactualprepared/list/ability/armoreligibility. V normalvoice,Stakesfreehand,Mhandaccess;focusonlyallowedfeature/noncostly/NONconsumedM,priced/consumedactualownedmaterials. Itemcasts206normallynoVSMunlessitemSAYSotherwise,scroll244itsOWNeligibility;neverblindlyimportclassprepared/materialgates. Positive:normalpricedMcannotfocus,qualifieditemexceptionseparate; negatives:missingV/S,untrainedwornarmor,classfocusforconsumedM,clientinventeditemexception. NOspell effects orclassgrants.', [68,273]),
 'cast-cost-privacy': ('PDF104-106 finite245sourceidentities: normaleligibleLslot>=spelllevel,ONLYONEspellslotTOCASTperTURN;OTHERslotexpenditures/conversionsNOTcount. AuthoritativelyPRIVATEINVALIDtargetconfirmedslotcastSPENDSSLOT/NOEFFECT;apparentsavefilterwhenwouldhavenoeffectonsuccess,otherwisenoeffect,NEVERreasonleak. UNVERIFIABLE/UNAUTHORIZEDfactsUNAVAILABLEwithnomutation. Positive:knownprivateinvalidHumanoidslotspendandfilteredreceipt; negatives:refundprivateinvalid,leakcreaturetype,banconversionafterslotcast,secondslotcast,unverifiedtargetcharged. Exactitem/free-usechannelsseparate,noinventedall-unknown-outcomeban.', [64,67,68]),
 'cast-long-ritual': ('PDF104-105 finite245sourceidentities: PREPARED+RitualTAGnormalcastingtimePLUS10min,NOslot,NOextraRitualAdept,NOritualUPCAST. Everycast>=1min usesMAGICACTION EACHturnandCASTINGConcentration;interruptionFAILSbutNOslotexpended,restartFROMBEGINNING. Positive:ComprehendLanguagesActionRitual10min+normalAction,LegendLore10minnormal; negatives:noPrepared,noTAG,omitted10min,ritualslotspend,ritualupcast,oneActionwholelongcast,interruptedslotspend. Existing906Detect/955Mendingretainexactscopes;thisnewadapterownsONLYsharedmodeprotocol.', [273,353,487]),
 'effect-time-span-stack': ('PDF106 finite245sourceidentities: ownONGOINGTIME-SPANspellCASTERmaydismissNOACTIONifNOTIncapacitated487,notownturnrestricted. SAMEspelloverlapsDONOTCOMBINE,strongestapplies,latestEQUAL;expired/endedsourcecannotreturn,olderstilllivecanresumeonnewerend. Positive:twoindependentHasteeffectsone+2AC,olderresumeswithoutclockreset; negatives:stack+4,clearallrecipientlinks,Incapacitateddismissal,Actioncost,foreigncaster,Instantaneousconsequencedismissal. Reuse351cast/recipientidentityAPIwithanewboundednonShieldFaithadapter,notduplicateitsoldbehavior.', [351,487]),
}

# Within-spell execution dependencies. First admission/manifest always precedes
# its downstream resolver; these extra edges bind multi-part output consumers.
EXTRA_DEPENDENCIES = {
 ('Arcane Hand','ending'): [('Arcane Hand','grasping-hand'),('Arcane Hand','interposing-hand')],
 ('Chromatic Orb','bounce'): [('Chromatic Orb','damage')],
 ('Cloudkill','damage-events'): [('Cloudkill','area-movement')],
 ('Detect Thoughts','probe'): [('Detect Thoughts','surface')],
 ('Flesh to Stone','petrification-ending'): [('Flesh to Stone','repeat-saves')],
 ('Freezing Sphere','delayed-globe'): [('Freezing Sphere','burst'),('Freezing Sphere','water')],
 ('Gaseous Form','ending'): [('Gaseous Form','movement-restrictions')],
 ('Prismatic Spray','special8'): [('Prismatic Spray','rays1-5'),('Prismatic Spray','indigo'),('Prismatic Spray','violet')],
 ('Sunburst','darkness-ending'): [('Darkness','effect')],
 ('Wall of Fire','movement-ending'): [('Wall of Fire','hot-side')],
}

CONFLICT_POLICY = {
 'sorcerer-mind-spike': {'class':'sorcerer','spell':'Mind Spike','list_page':68,'header_page':149},
 'sorcerer-phantasmal-force': {'class':'sorcerer','spell':'Phantasmal Force','list_page':68,'header_page':151},
 'wizard-phantasmal-force': {'class':'wizard','spell':'Phantasmal Force','list_page':80,'header_page':151},
}


SOURCE_PAGES = {'Alter Self': 107,
 'Animate Objects': 108,
 'Arcane Eye': 110,
 'Arcane Hand': 110,
 'Blindness/Deafness': 113,
 'Blink': 113,
 'Chain Lightning': 114,
 'Charm Monster': 114,
 'Charm Person': 115,
 'Chill Touch': 115,
 'Chromatic Orb': 115,
 'Circle of Death': 115,
 'Clairvoyance': 115,
 'Cloudkill': 116,
 'Color Spray': 116,
 'Comprehend Languages': 117,
 'Confusion': 117,
 'Counterspell': 120,
 'Creation': 121,
 'Dancing Lights': 121,
 'Darkness': 122,
 'Delayed Blast Fireball': 122,
 'Demiplane': 122,
 'Detect Thoughts': 123,
 'Dimension Door': 123,
 'Disguise Self': 124,
 'Disintegrate': 124,
 'Dominate Monster': 125,
 'Dominate Person': 126,
 'Dragon’s Breath': 126,
 'Earthquake': 127,
 'Enlarge/Reduce': 127,
 'Etherealness': 128,
 'Expeditious Retreat': 129,
 'Eyebite': 129,
 'False Life': 129,
 'Fear': 130,
 'Feather Fall': 130,
 'Finger of Death': 131,
 'Fire Shield': 132,
 'Fire Storm': 132,
 'Flame Blade': 132,
 'Flaming Sphere': 132,
 'Flesh to Stone': 133,
 'Fly': 133,
 'Freezing Sphere': 134,
 'Gaseous Form': 134,
 'Gate': 135,
 'Globe of Invulnerability': 136,
 'Grease': 137,
 'Greater Invisibility': 137,
 'Haste': 139,
 'Hold Monster': 140,
 'Hypnotic Pattern': 141,
 'Ice Knife': 141,
 'Incendiary Cloud': 142,
 'Invisibility': 143,
 'Knock': 143,
 'Legend Lore': 144,
 'Levitate': 144,
 'Mage Armor': 145,
 'Mage Hand': 145,
 'Magic Missile': 146,
 'Major Image': 147,
 'Mass Suggestion': 148,
 'Meteor Swarm': 149,
 'Minor Illusion': 149,
 'Mirror Image': 150,
 'Move Earth': 151,
 'Plane Shift': 153,
 'Power Word Stun': 154,
 'Prestidigitation': 154,
 'Prismatic Spray': 154,
 'Reverse Gravity': 159,
 'Scorching Ray': 159,
 'See Invisibility': 160,
 'Seeming': 160,
 'Shatter': 161,
 'Shield': 161,
 'Silent Image': 162,
 'Slow': 163,
 'Sorcerous Burst': 163,
 'Spider Climb': 164,
 'Suggestion': 166,
 'Sunbeam': 166,
 'Sunburst': 167,
 'Telekinesis': 167,
 'Teleport': 168,
 'Teleportation Circle': 168,
 'Thunderwave': 169,
 'Time Stop': 169,
 'Tongues': 169,
 'True Seeing': 171,
 'True Strike': 171,
 'Vampiric Touch': 171,
 'Vitriolic Sphere': 172,
 'Wall of Fire': 172}


REUSED_OUTPUTS = {'Acid Splash': [694],
 'Banishment': [939, 940],
 'Blight': [680],
 'Blur': [676],
 'Burning Hands': [677],
 'Command': [907],
 'Cone of Cold': [687],
 'Darkvision': [1046],
 'Daylight': [934],
 'Detect Magic': [906],
 'Dispel Magic': [355, 965, 966],
 'Dominate Beast': [1047, 1048],
 'Elementalism': [1051, 1052, 1053],
 'Enhance Ability': [1054],
 'Fire Bolt': [678],
 'Fireball': [679],
 'Fog Cloud': [682],
 'Gust of Wind': [1061],
 'Hold Person': [683],
 'Ice Storm': [686],
 'Insect Plague': [699],
 'Jump': [1064],
 'Light': [967],
 'Lightning Bolt': [691],
 'Magic Weapon': [924],
 'Mending': [955],
 'Message': [1070],
 'Misty Step': [688],
 'Poison Spray': [1075],
 'Polymorph': [698, 724, 725],
 'Power Word Kill': [506],
 'Protection from Energy': [1078],
 'Ray of Frost': [684],
 'Ray of Sickness': [695],
 'Shocking Grasp': [689],
 'Sleep': [690],
 'Sleet Storm': [685],
 'Stinking Cloud': [697],
 'Stoneskin': [1086],
 'Wall of Stone': [681, 722, 723],
 'Water Breathing': [1087],
 'Water Walk': [1088],
 'Web': [696, 726],
 'Wish': [583, 584, 585, 586, 587, 588, 589, 590, 591]}

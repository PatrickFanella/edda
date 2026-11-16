"""Finite pinned familiar source domain and independently readable predicates."""

NORMAL = {
 'Baboon': 'Small,AC12/HP3/Speed30Climb30;Bite+1/5ft/1d4-1Piercing;PackTactics allywithin5ftnotIncapacitated. Positiveally5ftAdvantage; rejectally6ft or Incapacitatedally.',
 'Badger': 'Tiny,AC11/HP5/Speed20Burrow5;PoisonResistance,Darkvision30,Perception3;Bite+2/5ft/1Piercing. PositivePoison10->5; rejectPoisonimmunity, Burrow20 or Bite1d4.',
 'Bat': 'Tiny,AC12/HP1/Speed5Fly30;Blindsight60,Passive11;Bite+4/5ft/1Piercing. PositiveBlindsight60; reject61ft,Fly60 or familiargenericattackpermission.',
 'Cat': 'Tiny,AC12/HP2/Speed40Climb40;DEXsave4,Perception3/Stealth4,Darkvision60;Scratch+4/5ft/1Slashing;JumperusesDEXnotSTR. PositiveDEX15jump; rejectSTR3jump or BitePiercing.',
 'Crab': 'Tiny,AC11/HP3/Speed20Swim20;Stealth2,Blindsight30;Amphibious;Claw+2/5ft/1Bludgeoning. Positiveairandwaterbreathing; rejectwateronly or Piercing.',
 'Deer': 'Medium,AC13/HP4/Speed50;Perception4,Darkvision60;AgilenoOpportunityAttackwhenmovingoutofenemyreach;Ram+2/5ft/1d4Bludgeoning. PositivewalkingexitnoOA; rejectonlyflyingexception or immunityallreactions.',
 'Eagle': 'Small,AC12/HP4/Speed10Fly60;Perception6/Passive16;Talons+4/5ft/1d4+2Slashing. PositiveD4=4->6; rejectFly50 or Piercing.',
 'Frog': 'Tiny,AC11/HP1/Speed20Swim20;Perception1/Stealth3,Darkvision30;Amphibious,StandingLeapLong10/High5withorwithoutrun;Bite+3/5ft/1Piercing. PositiveHigh5standing; rejectHigh10 or requires10ftrun.',
 'Giant Fire Beetle': 'Small,AC13/HP4/Speed30Climb30;FireResistance,Blindsight30;IlluminationBright10+Dim10;Bite+1/5ft/1Fire. PositiveFire10->5/20ftouterDim; rejectPoisonResistance,20ftBright or physicalBite.',
 'Goat': 'Medium,AC10/HP4/Speed40Climb30;STRsave2,Perception2,Darkvision60;Ram+2/5ft/1Bludgeoning OR1d4ifmoved20+ftstraighttowardimmediatelybeforehit. Positive20ft/D4=4->4; reject19fttrigger,Proneinferred or additional1+1d4.',
 'Hawk': 'Tiny,AC13/HP1/Speed10Fly60;Perception6/Passive16;Talons+5/5ft/1Slashing. Positivehit1; reject1d4,Fly50 or Blindsightinferred.',
 'Hyena': 'Medium,AC11/HP5/Speed50;Perception3,Darkvision60;PackTactics ally5ftnotIncapacitated;Bite+2/5ft/1d6Piercing. PositiveD6=6->6; rejectally6ftAdvantage or incapacitatedally.',
 'Jackal': 'Small,AC12/HP3/Speed40;Perception5/Stealth4,Darkvision90;Bite+1/5ft/1d4-1Piercing. PositiveD4=4->3; rejectDarkvision60 or PackTactics2014inferred.',
 'Lizard': 'Tiny,AC10/HP2/Speed20Climb20;Darkvision30;SpiderClimbdifficultsurfaces/ceilingsnoabilitycheck;Bite+2/5ft/1Piercing. Positiveceiling20; rejectcheckrequired or WebWalkerfromSpider.',
 'Octopus': 'Small,AC12/HP3/Speed5Swim30;Perception2/Stealth6,Darkvision30;Compression1inchnoextramovement,waterbreathingONLY;Tentacles+4/5ft/1Bludgeoning. Positive1inchpass; reject0.5inch,airbreathing or Grappleinferred. InkCloud separate.',
 'Owl': 'Tiny,AC11/HP1/Speed5Fly60;Perception5/Stealth5,Darkvision120;FlybynoOAwhenfliesoutofenemyreach;Talons+3/5ft/1Slashing. PositiveflyexitnoOA; rejectwalkexitexception,Darkvision60 or plan181closureasruntime.',
 'Piranha': 'Tiny,AC13/HP1/Speed5Swim40;Darkvision60;waterbreathingONLY;Bite+5/5ft/1Piercing,AdvantageiftargetnotallHP. Positive9of10HPAdvantage; rejectfullHPAdvantage,airbreathing orPackTactics.',
 'Rat': 'Tiny,AC10/HP1/Speed20Climb20;Perception2,Darkvision30;AgilenoOAwhenmovesoutenemyreach;Bite+2/5ft/1Piercing. PositivewalkexitnoOA; rejectFlybyonly,Fly20 or poisonBite.',
 'Raven': 'Tiny,AC12/HP2/Speed10Fly50;Perception3/Passive13;MimicryheardSIMPLEsounds,hearerWISInsightDC10discern;Beak+4/5ft/1Piercing. PositiveInsight10discerns; rejectnewunheardsound or automaticlanguagefluency.',
 'Scorpion': 'Tiny,AC11/HP1/Speed10;Blindsight10;Sting+2/5ft/1Piercing+1d6Poison. PositiveD6=6 yields1+6; rejectPoisonsave/Poisonedconditioninferred or Blindsight30.',
 'Seahorse': 'Tiny,AC12/HP1/Speed5Swim20;Perception2/Stealth5;waterbreathingONLY;CR0XP0;BubbleDashACTIONunderwatermoveSwimSpeedwithoutOA. Positive20ftunderwater; reject21ft,airBubbleDash or Biteattackinvented.',
 'Spider': 'Tiny,AC12/HP1/Speed20Climb20;Stealth4,Darkvision30;SpiderClimbceilingsnocheck;WebWalkerignoreswebmovementrestrictions/knowscreaturescontactingSAMEweb;Bite+4/5ft/1Piercing+1d4Poison. PositiveD4=4->1+4; rejectotherwebtracking,Poisonedcondition or genericwebcreation.',
 'Vulture': 'Medium,AC10/HP5/Speed10Fly50;Perception3/Passive13;PackTacticsally5ftnotIncapacitated;Beak+2/5ft/1d4Piercing. PositiveD4=4->4; rejectally6ftAdvantage or allyIncapacitated.',
 'Weasel': 'Tiny,AC13/HP1/Speed30Climb30;Acrobatics5/Perception3/Stealth5,Darkvision60;Bite+5/5ft/1Piercing. Positivehit1; rejectdamage1d4,Fly30 or readingStealth5asglobalinvisibility.',
}

FAMILIAR = {
 'familiar-cast-form': 'PDF130 Level1,1hourorRitual,10ft,VSM10GPincenseCONSUMED;unoccupiedspacechoose24pinnedCR0Beasts,spiritCelestial/Fey/FiendreplacingBeasttype. PositiveBat/Fey10GP; rejectunknownCR0withoutpinnedowner,CR1/8normalSnake,occupiedspace or10GPnonconsumption. ChainMagicActionmode separatelyconsumesbaseeffect.',
 'familiar-initiative-control': 'PDF130 familiarindependentally/obeyscommands,OWNInitiative/OWNturn,cannotattackbutothernormalqualifiedactions. PositiveDodge782ownturn; rejectcastersharedturn,normalBiteattackpermission,unqualifiedHideorallactionsinferred. Chainattackexceptionownsourceonly.',
 'familiar-telepathy': 'PDF130 within100fttelepathiccommunication;BonusActionsee/hearthroughsensesuntilSTARTyournextturn,benefitsactualspecialsenses. Positive100ftandqualifiedBatBlindsight60; reject101fttelepathy,freeBA,untilENDnextturn orprivatedisclosurewithoutknownsensoryfacts.',
 'familiar-touch-delivery': 'PDF130 touch-range spellcanbedeliveredbyfamiliarwithin100ftusingITSReactionwhenyoucast. Positive100ft/Reaction1; reject101ft,Reaction0,nonTouchspell,unqualifiedeffect or casterReactioncostinstead. Keepspellcastingmode/cost/targetauthority.',
 'familiar-dismiss-reappear': 'PDF130 at0HPdisappears,requirescastagain;MagicActiontemporarilydismisspocketdimensionorforever;MagicActionreappearfromtemporarydismissalunoccupiedwithin30ft. Positive30ft; reject31ft,normalreappearat0HPwithoutcast,freeaction oroccupiedspace. At0HP/pocketdisappearancewearing/carryingitemsleftoldspace.',
 'familiar-single-reform': 'PDF130 onlyONEfamiliar;recastwhileownedinsteadadoptsneweligibleform. PositiveBat->Owloneidentity/control; rejecttwofamiliars,newdicererollonreplayoroldformtraitspersist. Do notinfergenericcreaturehealing/gear/spawnsemantics.',
 'octopus-ink': 'PDF357–358/257 OctopusReaction1/Day:creatureendsownturnwithin5ftwhileunderwater;5ftCubeinkcenteredself,moveuptoSwimSpeed;HeavilyObscured1minuteorstrongcurrentdisperses. RecoverscompletedLongRestvia428,NOTdawn/calendarreset. Positive5fttrigger/move30; reject6ft,air,secondDailyuse,31ftmoveorunqualifiedcurrentfact.',
}

SPECIAL = {
 'Imp': {
  'statistics': 'TinyFiendDevilLawfulEvil,AC13/HP21/Speed20Fly40;STR6DEX17CON13INT11WIS12CHA14;Deception4Insight3Stealth5;ColdResistance,Fire/Poison/PoisonedImmunity,Darkvision120unimpededmagicalDarkness,Passive11,CommonInfernal,CR1PB2;MagicResistanceAdvantagevsSPELLS/magicaleffects. PositiveFire10->0Cold10->5; rejectLightningResistanceorgenericallsavesAdvantage.',
  'sting': 'Melee+5/5ft/1d6+3Piercing+2d6Poison. Positive6+3+12=21; rejectPoisonedcondition,DCpoisonsave or ordinaryfamiliarattackadmission.',
  'invisibility': 'ActioncastInvisibilitySELF,nocomponents,CHAability;consume330actualspellowner/962conditionprotocol. PositiveCHA2/freecomponents; rejectothers,implicitunbreakableInvisibility or copiedeffectowner.',
  'shape-shift': 'ActionratSpeed20/raven20Fly60/spider20Climb20/ortrueform;statsunchangedEXCEPTSpeed,gearNOTtransformed. PositiveImpHP21retainedasrat; rejectRatHP1,geartransformation orSpiderpoisonattack.',
 },
 'Pseudodragon': {
  'statistics': 'TinyDragonNeutralGood,AC14HP10Speed15Fly60;STR6DEX15CON13INT10WIS12CHA10;Perception5Stealth4;Blindsight10Darkvision60Passive15;understandsCommonDraconiccannotSPEAK;CR1/4PB2;MagicResistanceAdvantagevsmagic. PositiveBlindsight10; reject11ft,verbalspellcomponentsorallCONsavesAdvantage.',
  'bite-multiattack': 'BiteMelee+4/5ft/1d4+2Piercing;MultiattackTWO Bites,notSting. Positive2D4=4->12total; reject2Stings or interpretingChainONEattackasMultiattackaction.',
  'sting': 'ActionCONsaveDC12oneVISIBLEcreaturewithin5ft;failure2d4Poison+Poisoned1hour;failureby5+alsoUnconsciousWHILEPoisoned,endsdamageorcreaturewithin5ftACTIONwake. PositiveCON7bothconditions; rejectCON8Unconscious,successdamageorwakingendsPoisoned. Whole504/487/727conditions/endingsneeded.',
 },
 'Quasit': {
  'statistics': 'TinyFiendDemonChaoticEvil,AC13HP25Speed40;STR5DEX17CON10INT7WIS10CHA10;Stealth5;ColdFireLightningResistance,Poison/PoisonedImmunity;Darkvision120Passive10;AbyssalCommon;CR1PB2;MagicResistanceAdvantagevsmagic. PositiveFire10->5; rejectFireImmunity or speed20.',
  'rend': 'Melee+5/5ft/1d4+3SlashingandPoisoneduntilSTARTquasitnextturn. PositiveD4=4->7+boundPoisoned; rejectCONsaveinferred,untilENDtargetnextturn orPoisonedotherpacket.',
  'scare': 'Action1/DayWISDC10onecreaturewithin20ft;failureFrightened;repeatWISendeachownturn,successends;after1minuteautomaticsuccess. PDF257recoverscompletedLongRestvia428,NOTdawn/calendarreset. Positive20ftfailedDC10; reject21ft,secondDailyuse,foreverfearorignoringwhole488immunity/endings.',
  'invisibility': 'ActionInvisibilitySELF,nocomponents,CHA0;consume330sameeffectowner. Positivefreecomponents; rejectothercreaturecastorowncopyofeffect.',
  'shape-shift': 'Actionbat10Fly40/centipede40Climb40/toad40Swim40/ortrueform;otherstatsSAME,gearuntransformed. PositivebatHP25retained; rejectBatHP1 ornewstatblockattacks.',
 },
 'Skeleton': {
  'statistics': 'MediumUndeadLawfulEvilAC14HP13Speed30;STR10DEX16CON15INT6WIS8CHA5;BludgeoningVulnerability,Poison/Exhaustion/PoisonedImmunity;Shortbow/Shortswordgear;Darkvision60Passive9;understandsCommon+oneothercannotSPEAK;CR1/4PB2. PositiveBludgeoning5->10; rejectNecroticImmunity orverbalcasting.',
  'attacks': 'PDF325–326: source_page325isHEADERanchor; actualATTACKclausesPDF326. ShortswordMelee+5/5ft/1d6+3Piercing;ShortbowRanged+5/80/320ft/1d6+3Piercing. PositiveD6=6->9; rejectbow321ft,noDisadvantagelongrangeorMultiattackinvented. Genericstatistics/attacks ONLY; noAnimateDeadlifecycleorcommandpermission.',
 },
 'Sphinx of Wonder': {
  'statistics': 'TinyCelestialLawfulGoodAC13HP24Speed20Fly40;STR6DEX17CON13INT15WIS12CHA11;Arcana4Religion4Stealth5;NecroticPsychicRadiantResistance;Darkvision60Passive11;CelestialCommon;CR1PB2;MagicResistance. PositiveRadiant10->5; rejectForceResistanceormagicdamageimmunity.',
  'rend': 'Melee+5/5ft/1d4+3Slashing+2d6Radiant. PositiveD4=4and2D6=12->19; rejectNecroticbase,Poisonedcondition orordinaryfamiliarattacksadmitted.',
  'burst-ingenuity': 'Reaction2/DaytriggerSELFORothercreaturewithin30ftabilitycheckORsave:add2toroll. PDF257recoverscompletedLongRestvia428,NOTdawn/calendarreset. Positive30ftsave+2; reject31ft,attackroll,thirdDailyuse orrequiringknownoutcomeunnecessarily.',
 },
 'Sprite': {
  'statistics': 'TinyFeyNeutralGoodAC15HP10Speed10Fly40;STR3DEX18CON10INT14WIS13CHA11;Perception3Stealth8Passive13;CommonElvishSylvan;CR1/4PB2. PositiveStealth8; rejectDarkvisioninferredorHP2from2014.',
  'needle-sword': 'Melee+6/5ft/1d4+4Piercing. PositiveD4=4->8; reject+4attackor1fixedPiercing.',
  'enchanting-bow': 'Ranged+6/40/160ft:1PiercingANDCharmeduntilSTARTspritenextturn. Positivehit1+Charmed; rejectDCpoisonsave,2014Unconscious,161ftorendtargetnextturn;consume488wholeCharmedAPI.',
  'heart-sight': 'ActionCHAsaveDC10oneVISIBLEcreaturewithin5ft;CelestialFiendUndeadautomaticallyFAIL;failureknowemotion/alignment. PositiveFiendautosavefailure; reject6ft,humanoidsautofail,privatefactsexposedallplayersorunknownauthoritativealignmentguessed.',
  'invisibility': 'ActionInvisibilitySELF,nocomponents,CHA0;qualified330effectowner. Positive0components; rejecttargetothercreatureornewcopiedeffect.',
 },
 'Venomous Snake': {
  'statistics': 'TinyBeastUnalignedAC12HP5Speed30Swim30;STR2DEX15CON11INT1WIS10CHA3;Blindsight10Passive10;LanguagesNone;CR1/8PB2. PositiveChainformeligible; rejectnormalCR0formadmission,Blindsight30orspeaking.',
  'bite': 'Melee+4/5ft/1d4+2Piercing+1d6Poison. PositiveD4=4D6=6->12; rejectCONsaveorPoisonedconditioninferred.',
 },
}

# Create Undead's four named creature outputs. Zombie remains Sorcerer-owned;
# Wight Life Drain consumes that real output instead of copying its behavior.
UNDEAD = {
 'Ghoul': {
  'statistics': 'PDF288MediumUndeadChaoticEvilAC12HP22Speed30;STR13DEX15CON10INT7WIS10CHA6;Poison/Charmed/Exhaustion/PoisonedImmunity;Darkvision60Passive10/Common/CR1PB2. PositivePoison10->0; rejectNecroticImmunity orelftypechanges.',
  'bite-multiattack': 'PDF288–289BiteMelee+4/5ft/1d6+2Piercing+1d6Necrotic;MultiattackTWO Bites. PositivebothD6=6->14perhit; rejectClawreplacementwithoutsourceorallNecroticdamage.',
  'claw': 'PDF289Melee+4/5ft/1d4+2Slashing;nonUndeadNONelfcreatureCONDC10failureParalyzeduntilENDtargetnextturn. PositiveHumanoidnonelfD4=4->6+Paralyzed; rejectelfParalyzed,UndeadParalyzed orpoisonimmunitytreatedparalysisimmunity;whole504/727endings.',
 },
 'Ghast': {
  'statistics': 'PDF287–288MediumUndeadChaoticEvilAC13HP36Speed30;STR16DEX17CON10INT11WIS10CHA8;WISsave2;NecroticResistance,Poison/Charmed/Exhaustion/PoisonedImmunity;Darkvision60Passive10/Common/CR2PB2. PositiveNecrotic10->5; rejectNecroticImmunityorHP22.',
  'stench': 'PDF288startcreatureturnin5ftEmanationCONDC10failurePoisoneduntilSTARTtargetnextturn;successimmuneTHISghastStench24hours. Positive5ftfailedsave; reject6ft,entrytrigger,immunityallghastsor24hfromfailure;whole504/727immunity/ending.',
  'bite': 'PDF288Melee+5/5ft/1d8+3Piercing+2d8Necrotic. Positive3D8=8->27; rejectMultiattackinferredor2d6damage.',
  'claw': 'PDF288Melee+5/5ft/2d6+3Slashing;nonUndeadcreatureCONDC10failureParalyzeduntilENDtargetnextturn. PositiveelfNONUndeadeligible; rejectGhoulNONelfrestrictioncopied,UndeadParalyzedorDC13.',
 },
 'Wight': {
  'statistics-sunlight': 'PDF341MediumUndeadNeutralEvilAC14HP82Speed30;STR15DEX14CON16INT10WIS13CHA15;Perception3Stealth4,NecroticResistance,Poison/Exhaustion/PoisonedImmunity,StuddedLeatherArmor,Darkvision60Passive13,Common+oneother,CR3PB2;SunlightDisadvantagechecksANDattackrolls. PositiveSunlightattackDisadvantage; rejectsaveDisadvantageorNecroticImmunity.',
  'weapon-multiattack': 'PDF341NecroticSwordMelee+4/5ft/1d8+2Slashing+1d8Necrotic;Bow+4/150/600ft/1d8+2Piercing+1d8Necrotic. Multiattacktwoanycombination;mayreplaceONEwithLifeDrain. PositiveSword+Drain; rejecttwoDrains,thirdattack,601ftboworLifeDrainasattackroll.',
  'life-drain': 'PDF341CONsaveDC13onecreaturewithin5ftfailure1d8+2NecroticandreduceHPmaximumbyDAMAGETAKEN. PositiveD8=8/resistance->5maxloss; rejectsuccessdamage,unreducedroll10maxlossafterresistanceorunknownHPhealing/restorationinferred.',
  'zombie-rise': 'PDF341HumanoidslainbyTHISLifeDrainrises24hourslaterasZombieunderwightcontrolunlessrestoredtolife/bodydestroyed;max12controlled. Positive12thspawnat24h; reject13th,nonHumanoid,weaponkills,23h59m,revived/destroyedbody;consume330Zombieactualstat/action/Fortitudeoutputs,notnewcopy.',
 },
 'Mummy': {
  'statistics': 'PDF309MediumORSmallUndeadLawfulEvilAC11HP58Speed20;STR16DEX8CON15INT6WIS12CHA12/WISsave3;FireVulnerability;NecroticPoison/CharmedExhaustionFrightenedParalyzedPoisonedImmunity;Darkvision60Passive11/Common+twoother/CR3PB2. PositiveFire5->10; rejectFireResistanceorPsychicImmunity.',
  'fist-multiattack': 'PDF309RottingFistMelee+5/5ft/1d10+3Bludgeoning+3d6Necrotic;MultiattacktwoFistsANDDreadfulGlare. PositiveD10=10/3D6=18->31; rejectGlareinsteadoneFistoromittingcurseoncreature. Creature0HPfromTHISattackdiesandturnstodust.',
  'rot-curse': 'PDF309onFistcreaturehitcurse:noregainHP,HPmaximumdoesnotreturnnormalLongRest,HPmaximumdecreases3d6EACH24hours. Positive3sixes18maxlossday1; reject10fixeddespiteinjected18,ShortRestcure,healingwhilecursedorarbitrarybodydustat0MAXinsteadsourceattack0HP. Consume625/937exactcurseendinginterfaceswhereapplicable,notallcurseditemattunement.',
  'glare': 'PDF309WISDC11oneVISIBLEcreaturewithin60ftfailureFrighteneduntilENDmummynextturn;successimmuneTHISmummyGlare24h. Positive60ft; reject61ft,allmummiesimmunity,24hfromfailureor2014Paralyzedfeature;whole488required.',
 },
}


SOURCE_PAGES = {'Baboon': 345, 'Badger': 345, 'Bat': 345, 'Cat': 346, 'Crab': 347, 'Deer': 347, 'Eagle': 348, 'Frog': 348, 'Giant Fire Beetle': 351, 'Goat': 355, 'Hawk': 355, 'Hyena': 356, 'Jackal': 356, 'Lizard': 357, 'Octopus': 357, 'Owl': 358, 'Piranha': 358, 'Rat': 359, 'Raven': 359, 'Scorpion': 360, 'Seahorse': 361, 'Spider': 361, 'Vulture': 363, 'Weasel': 364, 'Imp': 300, 'Pseudodragon': 316, 'Quasit': 317, 'Skeleton': 325, 'Sphinx of Wonder': 327, 'Sprite': 329, 'Venomous Snake': 363, 'Ghoul': 288, 'Ghast': 287, 'Wight': 341, 'Mummy': 309}

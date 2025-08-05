基于这些数据实现， data/scripts/basic_tragedy_x.yaml 中的 script_models 字段， 该文件由 proto/tragedylooper/v1/script.proto 定义其结构, 如果ScriptModel缺少对应字段定义你添加

```json
[
    {
        "title": "Young Women’s Battlefield",
        "set": [
            {
                "name": "Tragedy Looper",
                "number": 3
            }
        ],
        "tragedySet": "basicTragedy",
        "daysPerLoop": 6,
        "difficultySets": [
            {
                "numberOfLoops": 4,
                "difficulty": 2
            },
            {
                "numberOfLoops": 3,
                "difficulty": 4
            }
        ],
        "mainPlot": [
            "signWithMe"
        ],
        "subPlots": [
            "loveAffair",
            "hiddenFreak"
        ],
        "cast": {
            "boyStudent": "person",
            "girlStudent": "friend",
            "classRep": "lovedOne",
            "shrineMaiden": "keyPerson",
            "policeOfficer": "person",
            "officeWorker": "lover",
            "informer": "serialKiller",
            "patient": "person",
            "nurse": "person"
        },
        "incidents": [
            {
                "day": 3,
                "incident": "foulEvil",
                "culprit": "officeWorker"
            },
            {
                "day": 4,
                "incident": "increasingUnease",
                "culprit": "classRep"
            },
            {
                "day": 6,
                "incident": "suicide",
                "culprit": "girlStudent"
            }
        ],
        "victory-conditions": "See Tragedy Looper Mastermind Handbook",
        "story": "See Tragedy Looper Mastermind Handbook",
        "mastermindHints": "See Tragedy Looper Mastermind Handbook"
    },
    {
        "title": "Magical Girls' Superiority",
        "creator": "M.Hydrome",
        "set": [
            {
                "name": "New Tragedies",
                "number": 3
            }
        ],
        "difficultySets": [
            {
                "numberOfLoops": 4,
                "difficulty": 2
            },
            {
                "numberOfLoops": 3,
                "difficulty": 3
            }
        ],
        "tragedySet": "basicTragedy",
        "mainPlot": [
            "signWithMe"
        ],
        "subPlots": [
            "unsettlingRumor",
            "hiddenFreak"
        ],
        "daysPerLoop": 5,
        "cast": {
            "popIdol": "keyPerson",
            "doctor": "conspiracyTheorist",
            "boyStudent": "serialKiller",
            "patient": "person",
            "teacher": "person",
            "richStudent": "person",
            "shrineMaiden": "friend"
        },
        "incidents": [
            {
                "day": 2,
                "incident": "increasingUnease",
                "culprit": "patient"
            },
            {
                "day": 3,
                "incident": "hospitalIncident",
                "culprit": "shrineMaiden"
            },
            {
                "day": 5,
                "incident": "missingPerson",
                "culprit": "boyStudent"
            }
        ],
        "specialRules": [
            ""
        ],
        "victory-conditions": "This is a script recommended for those who are playing :basicTragedy: for the first time. This is where it begins for real. Now you have 2 Subplots and seven people in the cast. Also, there are 5 days each loop\n\nThat said, the difficulty is still set to very low. Protagonists who come from :firstSteps: won’t feel super confused and will be able to enjoy the extra additions well enough. It’s not difficult to play as Mastermind either. But beware of the Final Guess. If all roles are clear, then you will lose, and that is more stressful than it may seem.",
        "story": "For boys and girls, the world is a constant battlefield. Each tiny little disturbance will shake their inexperienced souls. Especially in this town, where seductors lure with the promise of magic powers – without explaining the consequences\n\nThe school was bubbling with talk. “Spells” and “magic”, and everyone knows that you CAN become a magical girl, if you just sign the contract. Eventually, the :popIdol: will sign, and she will die. Can the Protagonists stop this tragedy from happening?",
        "mastermindHints": "The best way to win is through the Hospital :horror: and to place :intrigue: on the Idol. This reveals the least information.\n\nStart day 1 and 2 with :paranoia: +1 on the :patient:, and a bluff and :intrigue: +1 on the Shrine and Hospital respectively.\n\nAlso from day 1, use the :conspiracyTheorist: and trigger :spreading: :paranoia:. If you get some :intrigue: on the Hospital, put 2 :paranoia: on the :shrineMaiden:, and trigger the Hospital :horror:.\n\nOn day 3, you can use the Unsettling Rumor effect to put the second :intrigue: on the Hospital and kill the Protagonists.\n\nIf you can’t seem to get any :intrigue: on the Hospital, put instead 2 :intrigue: on the Idol and if possible, also on the Shrine, :richStudent:, and shrineMaiden. For this, :spreading: :paranoia: and :unsettlingRumor: will help you.\n\nHide the :serialKiller: as long as you can and use him to kill the :keyPerson: as your last resort. Killing the :friend: will make the Final Guess easy for the Protagonists, so avoid that as much as possible.\n\nThe Final Guess can easily become your downfall. Either you hide the :conspiracyTheorist: or the :friend:. Try to make it look like a :murder:er or a :brain: instead.\n\nVictory conditions\n1. At any day end, have the popIdol and the :boyStudent: alone in a location, triggering the :serialKiller: effect and thus the :keyPerson:’s loss condition immediately.\n\n2. At end of day 3, have at least 1 :intrigue: on the Hospital, at least 2 :paranoia: on the shrineMaiden, and the popIdol in the Hospital, triggering Hospital :horror:, killing the popIdol and triggering the :keyPerson:’s loss condition immediately.\n\n3. At loop end, have 2 :intrigue: on the popIdol, triggering the loss condition for “Sign with Me”.\n\n4. At any day end, have the shrineMaiden and the :boyStudent: alone in a location, triggering the :serialKiller: effect and thus the :friend:’s loss condition at loop end.\n\n5. At end of day 3, have at least 1 :intrigue: on the Hospital, at least 2 :paranoia: on the shrineMaiden who should be in the Hospital, triggering the Hospital :horror:, killing the shrineMaiden, and triggering the :friend:’s loss condition at loop end.\n\n6. At the end of day 3, have at least 2 :intrigue: on the Hospital, at least 2 :paranoia: on the shrineMaiden, triggering the Hospital :horror: killing the Protagonists."
    },
    {
        "title": "The Cat Box",
        "creator": "GEnd",
        "set": [
            {
                "name": "New Tragedies",
                "number": 4
            }
        ],
        "difficultySets": [
            {
                "numberOfLoops": 3,
                "difficulty": 5
            },
            {
                "numberOfLoops": 4,
                "difficulty": 3
            }
        ],
        "tragedySet": "basicTragedy",
        "mainPlot": [
            "giantTimeBomb"
        ],
        "subPlots": [
            "loveAffair",
            "unknownFactor"
        ],
        "daysPerLoop": 5,
        "cast": {
            "nurse": "witch",
            "alien": "lover",
            "officeWorker": "lovedOne",
            "blackCat": "factor",
            "richStudent": "person",
            "popIdol": "person",
            "soldier": "person"
        },
        "incidents": [
            {
                "day": 1,
                "incident": "suicide",
                "culprit": "richStudent"
            },
            {
                "day": 4,
                "incident": "missingPerson",
                "culprit": "officeWorker"
            },
            {
                "day": 5,
                "incident": "missingPerson",
                "culprit": "alien"
            }
        ],
        "description": "The Cat Box is a perfect script to play as your second :basicTragedy: script. The complexity is slightly higher, and there are now 8 characters in the cast. The idea of this script is to introduce two new things. \n\nThe first is the :mysteryBoy: and the :blackCat:; both a bit strange. It’s a good chance to get to know rather complicated characters. But don’t fret; neither of them will mess up things too badly. \n\nThe second is the trick to find the :witch:, who has mandatory :goodwill: Refusel. This will have the Protagonists learn how to use this technique.",
        "specialRules": [
            ""
        ],
        "victory-conditions": "1. At loop end, have 2 :intrigue: on the Hospital, triggering the loss condition of “:giantTimeBomb:.”\n\n2. At any day end, have 3 or more :paranoia: and 1 or more :intrigue: on the :officeWorker:, triggering the :lovedOne:'’s Protagonist kill.",
        "story": "The witch laughs and the cat familiar meows. A mistake in a ritual has summoned a strange existence from another world. And to make matters worse, it fell in love with a human. \n\nThe witch decides to burn the entire town to hide the evidence. But to do that, she needs to stay hidden. Time is of the essence: kill the witch and shut the lid on this box forever.",
        "mastermindHints": "You can win with either 2 :intrigue: on the Hospital or the ability of the :lovedOne:. Your first choice should be the former, but the latter hides more information. \n\nFirst day on loop 1 should be an :intrigue: +2 on the School, an :intrigue: +1 on the Hospital, and a bluff card on the Shrine. Then put an :paranoia: on the :richStudent: using the :conspiracyTheorist: and kill her with :suicide:. This is to give the impression that she might be the :witch: \n\nAfter that, try triggering both :missingPerson: Incidents by putting out :paranoia:. Use the :conspiracyTheorist:’s and :factor:’s abilities (Subplot “Unknown :factor:”: Putting 2 :intrigue: on the School makes the :blackCat: able to act as a :conspiracyTheorist:) when you need. Putting :paranoia: on the :officeWorker: is effective both for :missingPerson: and for the :lovedOne: ability. \n\nFinally, if you can get 2 :intrigue: on the School, Shrine, and Hospital, the first loop is perfected. You can try to get that result in future loops too, but it’ll be hard with Forbid :intrigue: and trying to prevent the Protagonists from finding the :witch: with :goodwill: Refusel. \n\nIt’s super easy to hide the :factor:, so you don’t need to worry about the Final Guess."
    }
]
```
```json
[
    {
        "title": "Crushed by the Hospital Building in Doronoki",
        "creator": "ロキルス",
        "set": [
            {
                "name": "New Tragedies",
                "number": 6
            }
        ],
        "difficultySets": [
            {
                "numberOfLoops": 5,
                "difficulty": 4
            },
            {
                "numberOfLoops": 4,
                "difficulty": 5
            }
        ],
        "tragedySet": "basicTragedy",
        "mainPlot": [
            "murderPlan"
        ],
        "subPlots": [
            "paranoiaVirus",
            "unknownFactor"
        ],
        "daysPerLoop": 6,
        "cast": {
            "informer": "keyPerson",
            "classRep": "killer",
            "boss": [
                "brain",
                {
                    "Turf": "Hospital"
                }
            ],
            "transferStudent": [
                "conspiracyTheorist",
                {
                    "enters on day": 5
                }
            ],
            "boyStudent": "factor",
            "shrineMaiden": "person",
            "alien": "person",
            "officeWorker": "person",
            "henchman": "person"
        },
        "incidents": [
            {
                "day": 2,
                "incident": "suicide",
                "culprit": "boyStudent"
            },
            {
                "day": 4,
                "incident": "hospitalIncident",
                "culprit": "henchman"
            },
            {
                "day": 5,
                "incident": "farawayMurder",
                "culprit": "transferStudent"
            }
        ],
        "description": "The special additions in this script are the presence of the powerful :boss: character and the capricious :transferStudent:. Have fun with the slightly higher complexity and the rather hard-to-use-well characters.",
        "specialRules": [
            ""
        ],
        "victory-conditions": "Note: Due to Subplot, :paranoiaVirus:, the :shrineMaiden:, :alien:, :officeWorker:, and :henchman: become :serialKiller:s while they have 3 or more :paranoia: on them. \n\n1. At any day end, have the Informant alone in a location with any of the :serialKiller:s, activating that ability, and triggering the loss condition for the :keyPerson: immediately. \n\n2. At any day end, have at least 2 :intrigue: on the Informant in the same location as the :classRep:, activating the :murder:er’s ability and thus triggering the :keyPerson:’s loss condition immediately. \n\n3. At the end of day 4, have the Informant in the Hospital, at least 1 :intrigue: on the Hospital, and at least 1 :paranoia: on the :henchman:, triggering the Hospital :horror:, killing the Informant, and triggering the loss condition of the :keyPerson: immediately. \n\n4. At the end of day 6, have at least 2 :paranoia: on the :transferStudent:, triggering the :farawayMurder:, and at least 2 :intrigue: on the Informant, choosing her to be the target, triggering the loss condition for the :keyPerson: immediately. \n\n5. Having at least 2 :intrigue: on the City, causing the :factor: (:boyStudent:) to gain the :keyPerson:’s effect, and applying any of (1) to (4) on the :factor: instead of the Informant. \n\n6. At the end of day 2, having at least 2 :intrigue: on the City, causing the :factor: to gain the :keyPerson:’s effect, and having at least 2 :paranoia: on the :boyStudent:, triggering :suicide:, and triggering the loss condition of the :keyPerson: immediately. \n\n7. At the end of day 4, have at least 2 :intrigue: on the Hospital and at least 1 :paranoia: on the :henchman:, triggering the Hospital :horror:’s Protagonist kill. \n\n8. At any day end, have at least 4 :intrigue: on the :classRep:, triggering the :murder:er’s Protagonist kill.",
        "story": "The hospital was empty, a ruin. Shut down by a combination of unfortunate events and political decisions. Forgotten, and abandoned. \n\nBut it harbored a secret. Here, in this very building, they had researched a new type of virus. One woman who worked as an informant had gotten ahold of this information, and the former CEO of the hospital, the :boss:, decided to silence her once and for all. Dedicated to his task, the :boss: plans to send his :henchman: to eliminate the evidence by destroying the building entirely, together with the boy who had contracted the virus. \n\nBut, one day a girl shows up, transferred back to the school in town. A girl with ties to the hospital and a burning hatred, accelerating the looming tragedy. \n\nCan the players outmaneuver the :henchman: and the vengeful girl?",
        "mastermindHints": "The :henchman: should start in the Hospital in every loop. \n\nThe first thing to aim for is a :factor: kill. First loop, :intrigue: +2 :intrigue: on the School, :intrigue: +1 on the City and a bluff on the Hospital. If the :boss: is in the City, use him to place the second :intrigue: on the City there, and :suicide: the :factor:. Use the same cards in Loop 2 but interchange them. \n\nIf that gets stopped, then you should aim for a :hospitalIorror:. The :boss: can spread :intrigue: here and there. The :goodwill: effects of the :henchman: and :alien: will get in your way. Spread out :paranoia: everywhere to make it hard to pinpoint the culprit. \n\nIf that also gets stopped, then it’s the :transferStudent:. With the :conspiracyTheorist:’s ability, you will for sure be able to activate :farawayMurder:, so have 2 :intrigue: on the :informant: (and :boyStudent:). \n\nWith the :paranoiaVirus:, you might activate a lot of :serialKiller:s:, but beware so you don’t get trapped in that. You can use that to get a win if you need. \n\nFor the Final Guess, you need to hide the :murder:er. Don’t use the :murder:er’s ability unless you can safely do so. "
    },
    {
        "title": "Those with Antibodies (NT)",
        "creator": "Satoru Sawamura",
        "set": [
            {
                "name": "New Tragedies",
                "number": 7
            }
        ],
        "difficultySets": [
            {
                "numberOfLoops": 5,
                "difficulty": 4
            },
            {
                "numberOfLoops": 4,
                "difficulty": 6
            }
        ],
        "tragedySet": "basicTragedy",
        "mainPlot": [
            "changeOfFuture"
        ],
        "subPlots": [
            "paranoiaVirus",
            "threadsFate"
        ],
        "daysPerLoop": 4,
        "cast": {
            "shrineMaiden": "cultist",
            "informer": "timeTraveler",
            "richStudent": "conspiracyTheorist",
            "classRep": "person",
            "officeWorker": "person",
            "forensicSpecialist": "person",
            "doctor": "person",
            "soldier": "person",
            "henchman": "person"
        },
        "incidents": [
            {
                "day": 1,
                "incident": "butterflyEffect",
                "culprit": "richStudent"
            },
            {
                "day": 2,
                "incident": "foulEvil",
                "culprit": "henchman"
            },
            {
                "day": 3,
                "incident": "spreading",
                "culprit": "doctor"
            },
            {
                "day": 4,
                "incident": "missingPerson",
                "culprit": "forensicSpecialist"
            }
        ],
        "description": "In each loop of this script, the Mastermind has a fool-proof way to win. That means that the Protagonists have no other choice than to find out what is happening and aim for the Final Guess. It’s a puzzle to solve. But not a regular, decent puzzle. And it’s a question of whether the Protagonists can notice this and find the answer.",
        "specialRules": [
            ""
        ],
        "victory-conditions": "* Note: Due to Subplot, :paranoiaVirus:, the :classRep:, :officeWorker:, :forensicSpecialist:, :doctor:, :soldier:, and :henchman: become :serialKiller:s while they have 3 or more :paranoia: on them. \n\n1. At the end of day 1, have at least 1 :paranoia: on the :richStudent:, triggering the :butterflyEffect:, thus triggering the loss condition in “Changing the Future” at loop end. \n\n2. End the final day with 2 or less :goodwill: tokens on the Informant, triggering the loss condition of the :timeTraveler:.",
        "story": "Cascading viruses—this is the future that we have no way to avoid. An insignificant start, small as the fluuttering of a butterfly’s wing, soon weaves its threads into a spiral, gradually turning it into an unavoidable maelstrom. The Protagonists cannot escape. They cannot change the future. Whatever plan of change they may have, it’s fruitless. Pointless. Completely in vain. \n\nThey must accept the future to live on and grasp the possibility of escaping the virus. Yes. That is the answer. They must find Those with Antibodies.",
        "mastermindHints": "The :henchman: should always start at the School.\n\nUnless you mess up badly, you should be able to win in every loop. In loop 1, put :paranoia: on :richStudent: and the :henchman:, and Forbid :goodwill: on the :forensicSpecialist:. Activate the :butterflyEffect:, if necessary, by using the :conspiracyTheorist:’s ability, and place a :goodwill: token on the :richStudent:. In this way, you’ll be able to trigger the :butterflyEffect: in all subsequent loops. \n\nAfter that, place :intrigue: tokens on the Shrine by the :foulEvil: effect, and delay the discovery of the Main Plot by, for example, placing :intrigue: on a girl. It’s recommended to use Forbid :goodwill: somewhere every single round. \n\nOne winning strategy for the Protagonists will be to try to use the effect from :paranoiaVirus:, that :person:s change into :serialKiller:s with enough :paranoia: counters, and by this method, discover which characters don’t change and which don’t die. :threadsFate: can also be used to pile up the :paranoia: tokens. So, once the Protagonists catch this one, make their lives miserable with :paranoia: -1 and Forbid :paranoia:."
    },
    {
        "title": "The Assassin from the Future",
        "creator": "unun",
        "set": [
            {
                "name": "Midnight Circle",
                "number": 3
            },
            {
                "name": "New Tragedies",
                "number": 5
            }
        ],
        "tragedySet": "basicTragedy",
        "mainPlot": [
            "changeOfFuture"
        ],
        "subPlots": [
            "hiddenFreak",
            "circleFriends"
        ],
        "difficultySets": [
            {
                "numberOfLoops": 5,
                "difficulty": 3
            },
            {
                "numberOfLoops": 4,
                "difficulty": 5
            }
        ],
        "daysPerLoop": 5,
        "cast": {
            "girlStudent": "person",
            "classRep": "friend",
            "shrineMaiden": "friend",
            "alien": "person",
            "informer": "conspiracyTheorist",
            "popIdol": "cultist",
            "journalist": "person",
            "doctor": "serialKiller",
            "patient": "timeTraveler"
        },
        "incidents": [
            {
                "day": 2,
                "incident": "butterflyEffect",
                "culprit": "classRep"
            },
            {
                "day": 4,
                "incident": "farawayMurder",
                "culprit": "doctor"
            },
            {
                "day": 5,
                "incident": "hospitalIncident",
                "culprit": "informer"
            }
        ],
        "victory-conditions": "See Tragedy Looper: Midnight Circle Mastermind Handbook",
        "story": "See Tragedy Looper: Midnight Circle Mastermind Handbook",
        "mastermindHints": "See Tragedy Looper: Midnight Circle Mastermind Handbook"
    },
    {
        "title": "Un Rerum",
        "creator": "GaRSoBaG",
        "set": [
            {
                "name": "New Tragedies",
                "number": 8
            }
        ],
        "difficultySets": [
            {
                "numberOfLoops": 4,
                "difficulty": 6
            }
        ],
        "tragedySet": "basicTragedy",
        "mainPlot": [
            "sealedItem"
        ],
        "subPlots": [
            "circleFriends",
            "threadsFate"
        ],
        "daysPerLoop": 6,
        "cast": {
            "ai": "brain",
            "richStudent": "cultist",
            "alien": "friend",
            "officeWorker": "friend",
            "blackCat": "conspiracyTheorist",
            "mysteryBoy": "lover",
            "classRep": "person",
            "forensicSpecialist": "person"
        },
        "incidents": [
            {
                "day": 1,
                "incident": "suicide",
                "culprit": "blackCat"
            },
            {
                "day": 2,
                "incident": "foulEvil",
                "culprit": "ai"
            },
            {
                "day": 3,
                "incident": "missingPerson",
                "culprit": "mysteryBoy"
            },
            {
                "day": 4,
                "incident": "spreading",
                "culprit": "officeWorker"
            },
            {
                "day": 5,
                "incident": "butterflyEffect",
                "culprit": "classRep"
            },
            {
                "day": 6,
                "incident": "murder",
                "culprit": "alien"
            }
        ],
        "description": "“un rerum”’s specialty is its usage of :threadsFate:. Horribly enough, if you fail in loop 3, you won’t be able to win the last loop. Also, the :ai: appears, and you’ll need to use that one’s abilities.",
        "specialRules": [
            ""
        ],
        "victory-conditions": "1. At loop end, have 2 :intrigue: on the Shrine, triggering the loss condition for “:sealedItem:”. \n\n2. At the end of day 6, have the :officeWorker: in the same location as the :alien:, and at least 2 :paranoia: on the :alien:, triggering Homicide, and thus triggering the loss condition of :friend: at loop end.",
        "story": "In a corner of the city, a lot of electronics have started blinking. One advanced computer, said to be able to change the world, has become self-aware. And it tries to replace reality with supernaturality, truly changing the world. The last hope IS the Shrine; that area has not yet fallen victim to this new order. If that one falls, then the entire city is lost",
        "mastermindHints": "Thanks to the :blackCat:, you can win by placing that needed :intrigue: on the Shrine. \n\nOn day 1, move the :richStudent: vertically, play :intrigue: +1 on the Shrine, and :intrigue: +2 on the City. If the Shrine’s :intrigue: is blocked, then you’ll have to use the :cultist:’s power to unblock it. After that, place :paranoia: tokens to trigger Transfer :friend:ship and :butterflyEffect:. \n\nIn loops 2 and 3, you’ll probably have :paranoia: on the :ai: and the :mysteryBoy:. If you trigger :foulEvil: or :missingPerson:, you’ll win. Specifically, in loop 2, the Protagonists will probably place :goodwill: tokens on the :mysteryBoy:, so you’ll mostly win by :missingPerson: in loop 3. \n\nIt’s also important to place :goodwill: on the :ai: and :mysteryBoy: with Transfer :friend:ship and :butterflyEffect:. If you start a loop with :paranoia: on both, place :intrigue: +2 on the :ai:, :intrigue: +1 on the Shrine, and :paranoia: +1 on the :mysteryBoy:, and you’ll have it. \n\nIf you avoid using :goodwill: Refusel on the :ai:, it’ll take time for the Protagonists to realize she’s the :brain:. Use the :goodwill: Refusel with wisdom. \n\nFor the Final Guess, hide either the :friend: or :conspiracyTheorist:."
    },
    {
        "title": "Lesser of Two Evils",
        "creator": "GEnd",
        "set": [
            {
                "name": "Tragedy Looper",
                "number": 4
            }
        ],
        "tragedySet": "basicTragedy",
        "daysPerLoop": 7,
        "difficultySets": [
            {
                "numberOfLoops": 4,
                "difficulty": 3
            },
            {
                "numberOfLoops": 3,
                "difficulty": 4
            }
        ],
        "mainPlot": [
            "sealedItem"
        ],
        "subPlots": [
            "hiddenFreak",
            "unknownFactor"
        ],
        "cast": {
            "boyStudent": "person",
            "girlStudent": "person",
            "richStudent": "brain",
            "shrineMaiden": "friend",
            "officeWorker": "serialKiller",
            "informer": "person",
            "journalist": "factor",
            "patient": "person",
            "nurse": "cultist"
        },
        "incidents": [
            {
                "day": 2,
                "incident": "increasingUnease",
                "culprit": "richStudent"
            },
            {
                "day": 4,
                "incident": "missingPerson",
                "culprit": "nurse"
            },
            {
                "day": 5,
                "incident": "missingPerson",
                "culprit": "boyStudent"
            },
            {
                "day": 7,
                "incident": "suicide",
                "culprit": "journalist"
            }
        ],
        "victory-conditions": "See Tragedy Looper Mastermind Handbook",
        "story": "See Tragedy Looper Mastermind Handbook",
        "mastermindHints": "See Tragedy Looper Mastermind Handbook"
    },
    {
        "title": "The Secret That Was Kept",
        "creator": "BakaFire",
        "set": [
            {
                "name": "Tragedy Looper",
                "number": 5
            }
        ],
        "tragedySet": "basicTragedy",
        "daysPerLoop": 7,
        "difficultySets": [
            {
                "numberOfLoops": 4,
                "difficulty": 3
            },
            {
                "numberOfLoops": 3,
                "difficulty": 5
            }
        ],
        "mainPlot": [
            "giantTimeBomb"
        ],
        "subPlots": [
            "threadsFate",
            "circleFriends"
        ],
        "cast": {
            "richStudent": "witch",
            "classRep": "person",
            "shrineMaiden": "person",
            "alien": "friend",
            "officeWorker": "friend",
            "informer": "conspiracyTheorist",
            "popIdol": "person",
            "journalist": "person",
            "patient": "person"
        },
        "incidents": [
            {
                "day": 2,
                "incident": "suicide",
                "culprit": "richStudent"
            },
            {
                "day": 3,
                "incident": "missingPerson",
                "culprit": "officeWorker"
            },
            {
                "day": 4,
                "incident": "hospitalIncident",
                "culprit": "journalist"
            },
            {
                "day": 6,
                "incident": "spreading",
                "culprit": "shrineMaiden"
            },
            {
                "day": 7,
                "incident": "foulEvil",
                "culprit": "popIdol"
            }
        ],
        "victory-conditions": "See Tragedy Looper Mastermind Handbook",
        "story": "See Tragedy Looper Mastermind Handbook",
        "mastermindHints": "See Tragedy Looper Mastermind Handbook"
    },
    {
        "title": "The Future of the Gods",
        "creator": "Nightly Moonfire group",
        "set": [
            {
                "name": "Tragedy Looper",
                "number": 6
            }
        ],
        "tragedySet": "basicTragedy",
        "daysPerLoop": 7,
        "difficultySets": [
            {
                "numberOfLoops": 4,
                "difficulty": 4
            }
        ],
        "mainPlot": [
            "changeOfFuture"
        ],
        "subPlots": [
            "hiddenFreak",
            "loveAffair"
        ],
        "cast": {
            "boyStudent": "timeTraveler",
            "richStudent": "person",
            "shrineMaiden": "cultist",
            "godlyBeing": [
                "lovedOne",
                {
                    "enters on loop": 3
                }
            ],
            "policeOfficer": "person",
            "officeWorker": "serialKiller",
            "popIdol": "lover",
            "patient": "person",
            "nurse": "friend"
        },
        "incidents": [
            {
                "day": 2,
                "incident": "suicide",
                "culprit": "popIdol"
            },
            {
                "day": 4,
                "incident": "increasingUnease",
                "culprit": "shrineMaiden"
            },
            {
                "day": 5,
                "incident": "butterflyEffect",
                "culprit": "policeOfficer"
            },
            {
                "day": 7,
                "incident": "foulEvil",
                "culprit": "patient"
            }
        ],
        "victory-conditions": "See Tragedy Looper Mastermind Handbook",
        "story": "See Tragedy Looper Mastermind Handbook",
        "mastermindHints": "See Tragedy Looper Mastermind Handbook"
    },
    {
        "title": "Mirror Passcode",
        "creator": "M. Hydromel",
        "set": [
            {
                "name": "Tragedy Looper",
                "number": 7
            }
        ],
        "tragedySet": "basicTragedy",
        "daysPerLoop": 7,
        "difficultySets": [
            {
                "numberOfLoops": 4,
                "difficulty": 4
            },
            {
                "numberOfLoops": 3,
                "difficulty": 6
            }
        ],
        "mainPlot": [
            "signWithMe"
        ],
        "subPlots": [
            "unknownFactor",
            "paranoiaVirus"
        ],
        "cast": {
            "boyStudent": "person",
            "girlStudent": "keyPerson",
            "richStudent": "factor",
            "mysteryBoy": "cultist",
            "officeWorker": "person",
            "informer": "conspiracyTheorist",
            "journalist": "person",
            "patient": "person",
            "nurse": "person"
        },
        "incidents": [
            {
                "day": 3,
                "incident": "missingPerson",
                "culprit": "richStudent"
            },
            {
                "day": 4,
                "incident": "increasingUnease",
                "culprit": "journalist"
            },
            {
                "day": 5,
                "incident": "hospitalIncident",
                "culprit": "mysteryBoy"
            },
            {
                "day": 7,
                "incident": "murder",
                "culprit": "boyStudent"
            }
        ],
        "victory-conditions": "See Tragedy Looper Mastermind Handbook",
        "story": "See Tragedy Looper Mastermind Handbook",
        "mastermindHints": "See Tragedy Looper Mastermind Handbook"
    },
    {
        "title": "Those with Antibodies",
        "creator": "Satoshi Sawamura",
        "set": [
            {
                "name": "Tragedy Looper",
                "number": 8
            }
        ],
        "tragedySet": "basicTragedy",
        "daysPerLoop": 4,
        "difficultySets": [
            {
                "numberOfLoops": 5,
                "difficulty": 5
            },
            {
                "numberOfLoops": 4,
                "difficulty": 6
            }
        ],
        "mainPlot": [
            "changeOfFuture"
        ],
        "subPlots": [
            "threadsFate",
            "paranoiaVirus"
        ],
        "cast": {
            "girlStudent": "person",
            "richStudent": "conspiracyTheorist",
            "classRep": "person",
            "shrineMaiden": "cultist",
            "policeOfficer": "person",
            "officeWorker": "person",
            "informer": "conspiracyTheorist",
            "doctor": "person",
            "patient": "person",
            "henchman": "timeTraveler"
        },
        "incidents": [
            {
                "day": 1,
                "incident": "butterflyEffect",
                "culprit": "richStudent"
            },
            {
                "day": 2,
                "incident": "foulEvil",
                "culprit": "henchman"
            },
            {
                "day": 3,
                "incident": "spreading",
                "culprit": "doctor"
            },
            {
                "day": 4,
                "incident": "missingPerson",
                "culprit": "policeOfficer"
            }
        ],
        "victory-conditions": "See Tragedy Looper Mastermind Handbook",
        "story": "See Tragedy Looper Mastermind Handbook",
        "mastermindHints": "See Tragedy Looper Mastermind Handbook"
    },
    {
        "title": "Prologue",
        "creator": "BakaFire",
        "set": [
            {
                "name": "Tragedy Looper",
                "number": 9
            },
            {
                "name": "New Tragedies",
                "number": 9
            }
        ],
        "tragedySet": "basicTragedy",
        "daysPerLoop": 7,
        "difficultySets": [
            {
                "numberOfLoops": 5,
                "difficulty": 5
            },
            {
                "numberOfLoops": 4,
                "difficulty": 7
            }
        ],
        "mainPlot": [
            "murderPlan"
        ],
        "subPlots": [
            "circleFriends",
            "loveAffair"
        ],
        "cast": {
            "boyStudent": "lover",
            "girlStudent": "lovedOne",
            "richStudent": "killer",
            "shrineMaiden": "keyPerson",
            "policeOfficer": "conspiracyTheorist",
            "officeWorker": "person",
            "informer": "friend",
            "doctor": "brain",
            "patient": "friend"
        },
        "incidents": [
            {
                "day": 2,
                "incident": "increasingUnease",
                "culprit": "officeWorker"
            },
            {
                "day": 4,
                "incident": "suicide",
                "culprit": "girlStudent"
            },
            {
                "day": 5,
                "incident": "hospitalIncident",
                "culprit": "boyStudent"
            },
            {
                "day": 7,
                "incident": "murder",
                "culprit": "policeOfficer"
            }
        ],
        "victory-conditions": "See Tragedy Looper Mastermind Handbook",
        "story": "See Tragedy Looper Mastermind Handbook",
        "mastermindHints": "See Tragedy Looper Mastermind Handbook"
    },
    {
        "title": "Neverending Happy & Sad Story",
        "creator": "BF + 3G",
        "set": [
            {
                "name": "Tragedy Looper",
                "number": 10
            },
            {
                "name": "New Tragedies",
                "number": 10
            }
        ],
        "tragedySet": "basicTragedy",
        "daysPerLoop": 6,
        "difficultySets": [
            {
                "numberOfLoops": 4,
                "difficulty": 8
            }
        ],
        "mainPlot": [
            "giantTimeBomb"
        ],
        "subPlots": [
            "unsettlingRumor",
            "loveAffair"
        ],
        "cast": {
            "girlStudent": "person",
            "richStudent": "lovedOne",
            "classRep": "person",
            "mysteryBoy": "brain",
            "alien": "person",
            "godlyBeing": [
                "witch",
                {
                    "enters on loop": 4
                }
            ],
            "officeWorker": "person",
            "popIdol": "person",
            "boss": [
                "conspiracyTheorist",
                {
                    "Turf": "School"
                }
            ],
            "patient": "lover",
            "nurse": "person"
        },
        "incidents": [
            {
                "day": 2,
                "incident": "butterflyEffect",
                "culprit": "classRep"
            },
            {
                "day": 3,
                "incident": "increasingUnease",
                "culprit": "alien"
            },
            {
                "day": 4,
                "incident": "missingPerson",
                "culprit": "officeWorker"
            },
            {
                "day": 5,
                "incident": "missingPerson",
                "culprit": "nurse"
            },
            {
                "day": 6,
                "incident": "missingPerson",
                "culprit": "patient"
            }
        ],
        "specialRules": [
            "Mastermind removes \"Forbid :goodwill:\" from his hand. It cannot be used in any loop"
        ],
        "victory-conditions": "See Tragedy Looper Mastermind Handbook",
        "story": "See Tragedy Looper Mastermind Handbook",
        "mastermindHints": "See Tragedy Looper Mastermind Handbook"
    },
    {
        "title": "The Illusion Under the Cherry Tree",
        "creator": "BakaFire",
        "set": [
            {
                "name": "New Tragedies",
                "number": 11
            }
        ],
        "difficultySets": [
            {
                "numberOfLoops": 4,
                "difficulty": 6
            }
        ],
        "tragedySet": "basicTragedy",
        "mainPlot": [
            "sealedItem"
        ],
        "subPlots": [
            "unsettlingRumor",
            "hiddenFreak"
        ],
        "daysPerLoop": 7,
        "cast": {
            "doctor": "brain",
            "boyStudent": "cultist",
            "godlyBeing": [
                "conspiracyTheorist",
                {
                    "enters on loop": 3
                }
            ],
            "richStudent": "serialKiller",
            "illusion": "friend",
            "shrineMaiden": "person",
            "officeWorker": "person",
            "informer": "person",
            "scientist": "person"
        },
        "incidents": [
            {
                "day": 2,
                "incident": "increasingUnease",
                "culprit": "shrineMaiden"
            },
            {
                "day": 4,
                "incident": "butterflyEffect",
                "culprit": "scientist"
            },
            {
                "day": 5,
                "incident": "missingPerson",
                "culprit": "illusion"
            },
            {
                "day": 7,
                "incident": "farawayMurder",
                "culprit": "richStudent"
            }
        ],
        "description": "This script is for players who are used to Tragedy Looper. The Protagonists should have at least 2 or 3 games under their belt. It circles around the uncanny abilities of the :illusion:. Enjoy the weirdness that changes the very point of many of the Action cards.",
        "specialRules": [
            ""
        ],
        "victory-conditions": "1. At loop end, have 2 :intrigue: on the Shrine, triggering the loss condition for “:sealedItem:”. \n\n2. At any day end, have the :richStudent: alone with the :illusion:, triggering the :serialKiller:’s ability and thus triggering the loss condition of the :friend: at loop end. \n\n3. At the end of day 7, have at least 1 :paranoia: on the :richStudent:, triggering :farawayMurder:, and at least 2 :intrigue: on the :illusion:, selecting her as a target, triggering the loss condition of the :friend: at loop end.",
        "story": "She has materialized in this world to protect the Seal at the Shrine. And by chance, she becomes friends to the Protagonists. The enemies are many: A fanatic who has lost his mind, am artistic killer, and an amnesiac god. The Protagonists must fight with her to protect the Seal, helped by her strange abilities. But the :illusion: will fade in the end. It will fade. :hope:fully, it will not do so in vain. And beautifully, under the Cherry Tree.",
        "mastermindHints": "You’ll aim for :intrigue: at the Shrine by using several methods, such as :unsettlingRumor:, :missingPerson:, and Action cards, as well as a killing of the :illusion: by the :serialKiller: or :farawayMurder:. It’s easiest if you don’t put all your eggs in one basket though. \n\nAs the :illusion: is there, the setting of Action cards on the board is much more powerful. For example, assume that you’ve set :intrigue: +1 on the Shrine and :paranoia: +1 on the School. If the :illusion: is in the School, you’ll get 1 :paranoia: on her, which looks good, as well as an unsettling :intrigue: on the Shrine. Try to use this as a way of confusing the Protagonists as much as you can. Bluntly put, you can put an Action card on the board where the :illusion: is, every single day. \n\nWinning by :farawayMurder: is a safe bet. The Protagonists will probably pump :goodwill: on the :illusion: to remove her from the board. You can’t completely stop that with Forbid :goodwill: but try to mess with that as much as you can. \n\nFor the Final Guess, hide the :conspiracyTheorist: or the :brain:. Once the :godlyBeing: enters in loop 3, it’s easy to hide either of them. "
    },
    {
        "title": "A Little Friend",
        "creator": "BakaFire+R",
        "set": [
            {
                "name": "New Tragedies",
                "number": 12
            },
            {
                "name": "Script Collection 2",
                "number": 1
            }
        ],
        "difficultySets": [
            {
                "numberOfLoops": 4,
                "difficulty": 4
            },
            {
                "numberOfLoops": 5,
                "difficulty": 3
            }
        ],
        "tragedySet": "basicTragedy",
        "mainPlot": [
            "changeOfFuture"
        ],
        "subPlots": [
            "hiddenFreak",
            "paranoiaVirus"
        ],
        "daysPerLoop": 5,
        "cast": {
            "godlyBeing": [
                "cultist",
                {
                    "enters on loop": 3
                }
            ],
            "journalist": "timeTraveler",
            "classRep": "serialKiller",
            "youngGirl": "friend",
            "richStudent": "conspiracyTheorist",
            "blackCat": "person",
            "alien": "person",
            "informer": "person",
            "forensicSpecialist": "person"
        },
        "incidents": [
            {
                "day": 1,
                "incident": "increasingUnease",
                "culprit": "richStudent"
            },
            {
                "day": 2,
                "incident": "missingPerson",
                "culprit": "youngGirl"
            },
            {
                "day": 4,
                "incident": "increasingUnease",
                "culprit": "informer"
            },
            {
                "day": 5,
                "incident": "butterflyEffect",
                "culprit": "classRep"
            }
        ],
        "description": "This is a script for players who have played at least once. It’s made to make full use of the abilities of the :youngGirl: and is a good script to get to know her usefulness.",
        "specialRules": [
            ""
        ],
        "victory-conditions": "* Note: Due to Subplot, :paranoiaVirus:, the :alien:, :blackCat:, Informant, and :forensicSpecialist: become :serialKiller:s while they have 3 or more :paranoia: on them. \n\n1. At the end of day 5, have at least 2 :paranoia: on the :classRep:, triggering the :butterflyEffect:, thus triggering the loss condition of “Changing the Future” at loop end. \n\n2. At loop end, have 2 or less :goodwill: on the :journalist:, triggering the loss condition of the :timeTraveler: at loop end. \n\n3. At any day end, have the :youngGirl: alone with the :classRep:, triggering the :serialKiller:’s ability, thus triggering the loss condition of the :friend: at loop end. \n\n4. At any day end, have at least 3 :paranoia: on :alien:, :blackCat:, Informant, or :forensicSpecialist:, triggering the effect of :paranoiaVirus:, turning them into :serialKiller:s, and alone with the :youngGirl:, triggering the loss condition of the :friend: at loop end.",
        "story": "There was a serial killer who threatened to kill a girl. A visitor from the future comes back to create killers to stop the first killer from succeeding. Things got out of control. \n\nThe Protagonists must reveal the visitor’s identity and persuade him to stop. And they will need the power of the young girl. She and only she can help them",
        "mastermindHints": "The first loop, aim to win by killing the :friend: (:youngGirl:). Put a Horizontal Movement on the :richStudent:, and :paranoia: +1 on the :youngGirl: and :classRep:. Then put an :paranoia: on the :richStudent: using the :conspiracyTheorist:’s ability, trigger :spreading: :paranoia:, and put two :paranoia: on the :classRep: and an :intrigue: on the :alien:. This way, you can kill the :friend: all the while hiding a lot of other information. If you succeed, just keep hiding information for all remaining loops. It’s important to keep the :classRep:’s :paranoia: at 3 or more. If you fail, keep going in loop 2 as per below, and try to hide the Main Plot. \n\nFrom loop 2 and on, keep killing the :friend:. If her power is used on day 1, use that to put :intrigue: on the Shrine with :missingPerson:. Generally, it’s easy to win, but the :serialKiller:s are hard to contain, and things are hard to hide. Try as much as you can to hide the :journalist:’s role of :timeTraveler: and try to keep the Main Plot hidden. It’s easy to camouflage the Main Plot as anything but Premeditated :murder:. Put :intrigue: on the Shrine and on the girls, to keep things fuzzy. \n\nFor the Final Guess, keep either Main Plot, the :paranoiaVirus: Subplot, who the first :serialKiller: is, or who the :timeTraveler: and the :conspiracyTheorist: are, hidden."
    },
    {
        "title": "Fall-Sakura Gathering",
        "creator": "Kyokei",
        "set": [
            {
                "name": "New Tragedies",
                "number": 13
            },
            {
                "name": "Script Collection 2",
                "number": 2
            }
        ],
        "difficultySets": [
            {
                "numberOfLoops": 4,
                "difficulty": 3
            },
            {
                "numberOfLoops": 5,
                "difficulty": 6
            }
        ],
        "tragedySet": "basicTragedy",
        "mainPlot": [
            "murderPlan"
        ],
        "subPlots": [
            "loveAffair",
            "hiddenFreak"
        ],
        "daysPerLoop": 5,
        "cast": {
            "sacredTree": "keyPerson",
            "soldier": "killer",
            "policeOfficer": "brain",
            "doctor": "lover",
            "classRep": "lovedOne",
            "shrineMaiden": "serialKiller",
            "sectFounder": "friend",
            "copycat": "brain",
            "mysteryBoy": "conspiracyTheorist"
        },
        "incidents": [
            {
                "day": 2,
                "incident": "suicide",
                "culprit": "classRep"
            },
            {
                "day": 4,
                "incident": "butterflyEffect",
                "culprit": "doctor"
            },
            {
                "day": 5,
                "incident": "increasingUnease",
                "culprit": "sectFounder"
            }
        ],
        "description": " This script is for players who have played at least once. It’s specifically designed to highlight the use three of the characters that originally came with the Second Script Collection: :copycat:, :sacredTree:, and :prophet:. ",
        "specialRules": [
            ""
        ],
        "victory-conditions": "1. At any day end, let the :shrineMaiden: be alone with the :sacredTree:, or have at least 2 :intrigue: on the :sacredTree: and in the same location as the :soldier: or :copycat:. This kills the :sacredTree: and triggers the loss condition for the :keyPerson: immediately. \n\n2. At any day end, let the :shrineMaiden: be alone with the :prophet:, triggering the :serialKiller:’s ability, and thus the loss condition for :friend: at loop end. \n\n3. At any day end, have at least 1 :intrigue: and at least 3 :paranoia: on the :classRep: (:lovedOne:), or at least 4 :intrigue: on the :soldier: or :copycat: (:murder:ers), triggering a Protagonist kill.",
        "story": "A giant cherry tree in the middle of the shrine. It has been used to keep evil powers away for centuries and is said to bloom in full even in fall. Hence, the worshipping of this tree has a long history in this little town, and for generations, this local religion has been passed down since the founder. But that's far from the truth. Long ago, a warlord hid a large sum of money under the tree and died, and this religious lie has been kept alive to protect the treasure. And some individuals, who have found out this truth, are coming to town. They have fooled the shrine maiden and made her insane. They lured a young boy into town and pushed him down the path of evil. Their goal is to cut down the cherry tree and take the treasure for themselves. The Protagonists must find out their scheme and protect the secret. ",
        "mastermindHints": "In loop 1, put a Vertical Movement on the :prophet: and end everything straight away. The remaining two actions can be Forbid :goodwill: and Horizontal Movement on the :copycat:. If you by some reason should fail to end it there, put some :paranoia: on the :prophet: with the :conspiracyTheorist:’s ability. \n\nFrom loop 2 and on, start aiming for a :lovedOne: victory. However, for that, the :suicide: on day 2 is a crutch. The best is if :classRep: has exactly 1 :paranoia: on day 2. \n\nThe :spreading: :paranoia: of the :prophet: has extreme explosive power. If the :classRep: lives, put 1 :intrigue: and 1 :paranoia: on her, and you’ve basically won. Or you could put 2 :intrigue: on the :sacredTree: and kill it with the :murder:er. It should be easy if you use the :conspiracyTheorist:’s ability. \n\nIt looks as if it’s easy to win, but the roles are as easily revealed too, so take great care. If you put cards on the two :murder:ers, and total 4 :intrigue: on them, you can win on pure luck, but you’ll end up revealing both :murder:ers. If you put :intrigue: on the :sacredTree: outside of :spreading: :paranoia:, you won’t have time for much more, and it’s hard to win by just that. \n\nIt’s next to hopeless to hide much for the Final Guess. You need to stop the deaths of the :friend:, :lovedOne:, and :lover:, and try to desperately cloud who are the :murder:ers and :brain:. "
    },
    {
        "title": "Thunder in the City",
        "creator": "Redless",
        "difficultySets": [
            {
                "numberOfLoops": 5,
                "difficulty": 0
            }
        ],
        "tragedySet": "firstSteps",
        "mainPlot": [
            "lightAvenger"
        ],
        "subPlots": [
            "unsettlingRumor"
        ],
        "daysPerLoop": 5,
        "cast": {
            "popIdol": "brain",
            "girlStudent": "conspiracyTheorist",
            "journalist": "person",
            "patient": "person",
            "shrineMaiden": "person",
            "boyStudent": "person",
            "richStudent": "person"
        },
        "incidents": [
            {
                "day": 3,
                "incident": "hospitalIncident",
                "culprit": "boyStudent"
            },
            {
                "day": 4,
                "incident": "murder",
                "culprit": "richStudent"
            },
            {
                "day": 5,
                "incident": "missingPerson",
                "culprit": "shrineMaiden"
            }
        ],
        "specialRules": [
            ""
        ],
        "victory-conditions": "",
        "story": "",
        "mastermindHints": "This script is in an awkward spot, since it requires a fair bit of skill to solve and because it's very long, yet it's a :firstSteps: script. the kind of people who would like this script would definitely be good at this game, and if that's the case, why play a :firststeps: script with them? maybe if some people haven't played it before and some have? anyways, this script is meant to introduce the players to deep bluffing and trickery, 'coverups' in other words. to do that, we have the :brain: and :increasingUnease: subplot to obfuscate who the :brain: is. loop 1, you should try to place an :intrigue: on the city, school, and hospital. I'd put the 2 on the city since it's most important and they're less likely to block it, and one on the hospital to make it easier to set the incident off (if you definitely don't want the incident, you can also do school). then, you want to set off the :hospitalIncident: by placing a :paranoia: AND using the contheorist's ability on the :boyStudent: on day 3, to surprise the protagonists, while placing :paranoia: on other people to bluff the culprits. alternately, if they blocked on hospital or something, try to get them to block there again for the free school :intrigue:, move the :journalist: or :popIdol: over (:journalist: probably better) and/or trigger :missingPerson: so that you can go for the :placeProtect: bluff. in future loops, try to confuse the protagonists about who and where the :brain: is, and who does the :hospitalIncident:. the presence of the :murder: incident makes it harder to bluff that RMD does the :hospitalIncident:, but makes it easier to bluff place to protect and ultimately a hideous script. slowly, more and more things will be revealed over the course of five loops. it's subtle, it's a marathon, and it requires GOOD play from the protagonists to have any chance at all, but I hope and think it does a suitable job introducing the more trickery focused side of looper."
    },
    {
        "title": "A Cruel Shrine Maiden's Thesis",
        "creator": "Redless",
        "difficultySets": [
            {
                "numberOfLoops": 3,
                "difficulty": 0
            },
            {
                "numberOfLoops": 4,
                "difficulty": 0
            }
        ],
        "tragedySet": "firstSteps",
        "mainPlot": [
            "placeProtect"
        ],
        "subPlots": [
            "shadowRipper"
        ],
        "daysPerLoop": 4,
        "cast": {
            "nurse": "keyPerson",
            "popIdol": "cultist",
            "girlStudent": "conspiracyTheorist",
            "shrineMaiden": "serialKiller",
            "classRep": "person",
            "patient": "person",
            "officeWorker": "person"
        },
        "incidents": [
            {
                "day": 1,
                "incident": "increasingUnease",
                "culprit": "officeWorker"
            },
            {
                "day": 4,
                "incident": "missingPerson",
                "culprit": "nurse"
            }
        ],
        "specialRules": [
            ""
        ],
        "victory-conditions": "",
        "story": "",
        "mastermindHints": "This whole script revolves around :missingPerson:. It will let you drag the :nurse: into the maiden's path, and also let you stack some more :intrigue: on the school. It's your \"easy out\", everything else in the script lives in service of it. The first day you should aim to serial kill the nurse. LOL Gotem. Another loop, drop :paranoia: on the OW, and move the :girlStudent: left. By doing this, you will be in a good spot to MP. If people don't die, stack the school and set to work adding :intrigue:. If people do die, try to position the :shrineMaiden: alone so that you can MP the nurse over. Another loop, try and stack people in the school and obfuscate who exactly the :cultist: is. Finally, try and win the final loop via :missingPerson: shenanigans, which is surprisingly hard to do. You might consider playing this instead of the first script, or in place of the second script if your players don't mind you crushing their spirits by doing the same BS strat that characterizes the first script, a second time in a row. I believe this is very balanced. But of course you have to contend with the repetition of :serialKiller: if you're playing this after the first script. "
    },
    {
        "title": "Trouble in Paradise",
        "creator": "chewonki",
        "difficultySets": [
            {
                "numberOfLoops": 4,
                "difficulty": 0
            },
            {
                "numberOfLoops": 5,
                "difficulty": 0
            }
        ],
        "tragedySet": "basicTragedy",
        "mainPlot": [
            "signWithMe"
        ],
        "subPlots": [
            "hiddenFreak",
            "loveAffair"
        ],
        "daysPerLoop": 4,
        "cast": {
            "richStudent": "keyPerson",
            "alien": "serialKiller",
            "godlyBeing": [
                "friend",
                {
                    "enters on loop": 3
                }
            ],
            "henchman": "lover",
            "patient": "lovedOne",
            "mysteryBoy": "person",
            "classRep": "person",
            "doctor": "person",
            "popIdol": "person",
            "informer": "person"
        },
        "incidents": [
            {
                "day": 2,
                "incident": "increasingUnease",
                "culprit": "richStudent"
            },
            {
                "day": 3,
                "incident": "foulEvil",
                "culprit": "patient"
            },
            {
                "day": 4,
                "incident": "suicide",
                "culprit": "henchman"
            }
        ],
        "description": "",
        "specialRules": [
            ""
        ],
        "victory-conditions": "",
        "story": "",
        "mastermindHints": ""
    },
    {
        "title": "Tough to Name",
        "creator": "Hal",
        "difficultySets": [
            {
                "numberOfLoops": 3,
                "difficulty": 0
            },
            {
                "numberOfLoops": 4,
                "difficulty": 0
            }
        ],
        "tragedySet": "basicTragedy",
        "mainPlot": [
            "giantTimeBomb"
        ],
        "subPlots": [
            "loveAffair",
            "hiddenFreak"
        ],
        "daysPerLoop": 6,
        "cast": {
            "popIdol": "witch",
            "patient": "lover",
            "boyStudent": "lovedOne",
            "shrineMaiden": "serialKiller",
            "godlyBeing": [
                "friend",
                {
                    "enters on loop": 2
                }
            ],
            "nurse": "person",
            "journalist": "person",
            "officeWorker": "person",
            "richStudent": "person",
            "classRep": "person"
        },
        "incidents": [
            {
                "day": 2,
                "incident": "missingPerson",
                "culprit": "boyStudent"
            },
            {
                "day": 4,
                "incident": "increasingUnease",
                "culprit": "journalist"
            },
            {
                "day": 5,
                "incident": "hospitalIncident",
                "culprit": "shrineMaiden"
            },
            {
                "day": 6,
                "incident": "suicide",
                "culprit": "patient"
            }
        ],
        "description": "",
        "specialRules": [
            ""
        ],
        "victory-conditions": "",
        "story": "",
        "mastermindHints": "The witch might get discovered quite early - the :popIdol:'s goodwill ability is a tempting one, so her goodwill refusel will probably be outed in the first loop. Given this, Max (our mastermind) decided to try and win through intrigue on the City in day one, and spread about some intrigue and paranoia elsewhere so it wasn't too obvious.\n\nThe hospital incident is another way for the mastermind to win, especially coupled with increasing unease. But you don't have many ways of getting intrigue out, and no cultist to block you forbid intrigues. The only one you do have is :missingPerson:, and that's quite one use as the culprit reveals themselves. Once the :godlyBeing: is in play, the players could stack goodwill on them too to get the ability to remove intrigue, and to discover the culprit. So the hospital incident looks quite tough to pull off. On the plus side, the culprit is the serial killer, so she won't get offed. It did trigger with one intrigue in the hospital in loop 2 for us, which isn't great for the mastermind, as it might reveal the :patient: as the lover. It killed the patient and the boy student, which was a happy coincidence for Max.\n\nThe :shrineMaiden: serial killer is a good secret weapon for loop 2, if you can avoid her being discovered. The :godlyBeing: appears in the shrine in loop 2, so if the protagonists don't move the two apart, you can kill the friend and spend the rest of the loop misleading them.\n\nI'm not sure what the best role to try and hide is for the mastermind. My suspicion is that it's the lover/loved one. You have the option of pulling a win out of the bag on the last day using the suicide on the patient, but it looks like you might have trouble surviving the final guess if you did that. Maybe you can get away with hiding the :witch:, but you'd have to be lucky I think."
    },
    {
        "title": "Was it Just the Wind?",
        "creator": "Tuxian",
        "difficultySets": [
            {
                "numberOfLoops": 4,
                "difficulty": 4
            },
            {
                "numberOfLoops": 3,
                "difficulty": 6
            }
        ],
        "tragedySet": "basicTragedy",
        "mainPlot": [
            "murderPlan"
        ],
        "subPlots": [
            "threadsFate",
            "unsettlingRumor"
        ],
        "daysPerLoop": 7,
        "cast": {
            "illusion": "keyPerson",
            "alien": "killer",
            "journalist": "brain",
            "shrineMaiden": "conspiracyTheorist",
            "patient": "person",
            "doctor": "person",
            "popIdol": "person",
            "richStudent": "person",
            "classRep": "person",
            "henchman": "person",
            "mysteryBoy": "factor"
        },
        "incidents": [
            {
                "day": 1,
                "incident": "butterflyEffect",
                "culprit": "richStudent"
            },
            {
                "day": 2,
                "incident": "missingPerson",
                "culprit": "shrineMaiden"
            },
            {
                "day": 3,
                "incident": "increasingUnease",
                "culprit": "henchman"
            },
            {
                "day": 4,
                "incident": "increasingUnease",
                "culprit": "alien"
            },
            {
                "day": 5,
                "incident": "farawayMurder",
                "culprit": "patient"
            },
            {
                "day": 6,
                "incident": "spreading",
                "culprit": "doctor"
            },
            {
                "day": 7,
                "incident": "suicide",
                "culprit": "mysteryBoy"
            }
        ],
        "description": "",
        "specialRules": [
            ""
        ],
        "victory-conditions": "",
        "story": "",
        "mastermindHints": "Loop 1, activate the :factor: as fast as possible, bluff Incidents to keep the Protagonists guessing and putting out fires. Kill :factor: with either :farawayMurder: (preferred) or :suicide:. Ignore the :illusion: to make them ignore it as well.\n\nLoop 2, attempt to kill the :illusion: using the :killer:. Gather several people together to hide who the :killer: might be.\n\nLoop 3, use :farawayMurder: to kill the :illusion: or activated :factor:, or :suicide: the :factor:.\n\nLoop 4, if playing, use whatever you need to.\n\nHide the :brain: by using the Unsettling Rumor, and hide the :killer: by having a crowd around the :illusion:. Remember the :illusion: can be used to bluff a different main plot because she gets tokens when you play a card on a Location, you can also use this to your advantage by bluffing something where she is and moving her to a location where she would get :intrigue:.\n\n[BGG Log 1](https://boardgamegeek.com/thread/1548338) and [BGG Log 2](https://boardgamegeek.com/thread/1550297) are two logs of this script played where the Mastermind wins, so you can study those if you’re trying to optimize "
    },
    {
        "title": "Tofu Murder Case",
        "creator": "じーちゃん",
        "source": "https://web.archive.org/web/20171025123847/http://g-chan.dip.jp/square/archives/2015/12/rooper_x_02.html",
        "difficultySets": [
            {
                "numberOfLoops": 4,
                "difficulty": 0
            }
        ],
        "tragedySet": "firstSteps",
        "mainPlot": [
            "murderPlan"
        ],
        "subPlots": [
            "hideousScript"
        ],
        "daysPerLoop": 5,
        "cast": {
            "popIdol": "keyPerson",
            "girlStudent": "killer",
            "policeOfficer": "brain",
            "richStudent": "conspiracyTheorist",
            "patient": "curmudgeon",
            "shrineMaiden": "friend",
            "boyStudent": "person"
        },
        "incidents": [
            {
                "day": 2,
                "incident": "suicide",
                "culprit": "shrineMaiden"
            },
            {
                "day": 3,
                "incident": "murder",
                "culprit": "richStudent"
            },
            {
                "day": 5,
                "incident": "increasingUnease",
                "culprit": "patient"
            }
        ],
        "description": "This :firstSteps: scenario was made for players who have never played Tragedy Looper before. However, it's not particularly easy in terms of difficulty. In fact, it's somewhat difficult. I think that even newcomers can enjoy it, but I would not recommend it to people who aren't good at logic puzzles, or if they're new to board games in general (well, in this case, they probably won't be playing Tragedy Looper right now to begin with). This scenario is intended to be played with table talk allowed, but you can forbid it if the players are experienced. Since a tofu-minded lady commits murders in this scenario, I titled it \"Tofu :murder: Case.\"\n\nThis scenario consists of four Loops, but the first three Loops will allow you to experienced the core element of Tragedy Looper—information gathering. The main focus of this is to learn how to deduce clues based on what's happening in the board, such as \\\"The board was like this when we died, so this must be the Plot.\\\"\n\nBefore entering the final Loop, the Mastermind will defeat the Protagonists while guiding them to reveal the Main Plot, Sub Plot, all Roles, and all Incident Culprits (Since :firstSteps: doesn't have a Final Guess, you won't lost just for revealing everything).\n\nThen, in the final Loop, the Protagonists will have a guessing game with the Mastermind based on the information that's been revealed. The aim is for the Protagonists to play in a way that avoids the Defeat Condition when they have all the information at hand.",
        "specialRules": [
            ""
        ],
        "victory-conditions": "1. Kill the :keyPerson: (:killer:, :murder:)\n2. Kill the :friend: (:suicide:, :murder:)\n3. Kill the Protagonists (:killer:)",
        "story": "",
        "mastermindHints": "First Loop\n\nThe goal is the won by putting 4 :intrigue: on the :killer:, which is the most difficult option. Since you can get 1 :intrigue: by triggering :increasingUnease: on the 5th Day, you only need to put 3 :intrigue: on the :girlStudent:.\n\nDay 1\n- Pop Idol: :intrigue:+1\n- Patient: :paranoia:+1\n- :girlStudent Student: :paranoia:+1 (this is a bluff, so you can use others if you want)\n\nPlace :paranoia:+1 on the :patient: every turn to trigger the :increasingUnease: Incident. The Protagonists have a total of 3 :paranoia:-1 cards, so you can reach the :paranoia: limit on the final Day if you put :paranoia:+1 out every turn. \n\nHowever, this time, there is a :popIdol:. Since there is a possibility of her removing :paranoia: via her mandatory, if she accumulates :goodwill: on her, you can counter it with repeated Forbid :goodwill:, preventing her from moving to the Hospital, or simply place a 4th :intrigue:. Additionally, placing it on :popIdol:, :girlStudent:, and :patient: is also a setup for the next Loop, so it's recommended to place it on these three Characters.\n\nDay 2\n- :girlStudent Student: :intrigue:+1\n- Patient: :paranoia:+1\n- Pop Idol: :paranoia:+1 (this is a bluff, so you can use others if you want)\n\nContinuing from the previous Day, place :paranoia: on the :patient:. Then place :intrigue:+1 on the :girlStudent:. This is the day when :suicide: occurs, but if the Mastermind does not place any :paranoia:, the :shrineMaiden: shouldn't hit her required limit.\n\nDay 3\n- :girlStudent Student: :intrigue:+2\n- Patient: :paranoia:+1\n- Pop Idol: Move to the same position as :richStudent:\n\nAt this point, it's important to place :intrigue:+2 while moving the :richStudent:. If :intrigue:+2 is blocked, your options for beating the Protagonists will be greatly narrowed (I think the only option then is to tput 2 :intrigue: on the :keyPerson: and kill him with the :killer:).\n\nAs insurance in case :intrigue:+2 is nullified, move :popIdol: (:keyPerson:) towards :richStudent: (:conspiracyTheorist:). By doing this, the Daughter, who is a :conspiracyTheorist:, will be abl to put a Paranoid counter on herself, trigger :murder:, kill the :keyPerson:, and defeat the Protagonists.\n\nIf :intrigue:+2 is successfully placed on the :girlStudent: (:killer:), don't use :conspiracyTheorist:'s mandatory ability and continue instead with \\\"No Incident occured.\\\"\n\nDay 4\n- :girlStudent Student: :intrigue:+1\n- Patient: :paranoia:+1\n- Pop Idol: :paranoia:+1 (this is a bluff, so you can use others if you want)\n\nWhen the :girlStudent: gets 4 :intrigue:, kill the Protagonists to end the Loop. However, the Protagonists will likely be wary of her since she has three :intrigue:, so they will likely place Forbid :intrigue:. It's a good idea to place :paranoia: so that the :popIdol: also reaches her Limit, so it makes it difficult to determine if the final Day's Culprit is the :popIdol: or the :patient:.\n\nDay 5\n- :girlStudent Student: :intrigue:+1\n- Patient: :paranoia:+1\n- Pop Idol: :paranoia:+1 (this is a bluff, so you can use others if you want)\n\nSame as Day 4. If everything goes according to plan, the :girlStudent: should have 3 :intrigue: on her. Now, trigger :increasingUnease: (which should be unstoppable) to put the 4th :intrigue: on the :killer:, and kill the Protagonists at the end of the turn to end the Loop.\n\n2nd Loop\n\nThe strategy for the 2nd Loop is to aim for a one-turn kill, or kill the :keyPerson: with the :killer: if you can't get a one-turn kill.\n\nDay 1\n- Pop Idol: :intrigue:+2\n- :girlStudent Student: Move to the City\n- Patient: :paranoia:+1\n\nLet's aim for a one-turn kill\nSince you're placing cards on the same characters as in the previous Loop, and, according to the plot, the Protagonists lost the previous Loop due to the :killer:'s mandatory :intrigue: 4 ability, the Protagonists will probably place Forbid :intrigue: on the :girlStudent: (:killer:). Well, they may move the :popIdol: to avoid the risk of a one-turn kill, but if they, you can chase them down with the :killer: the next Day or later.\n\nDay 2 onward\n\nSo, did you get the one-turn kill? Well, you're reading this on Day 2, so probably not (lol). From hereon, use :policeOfficer: (:brain:)'s mandatory ability aggressively and aim to put :intrigue: on the :keyPerson: and :killer:.\n\nThere's no Final Guess in the First Step set, so there's no reason to be stingy with mandatory role abilities. It's also a good idea to aim to kill the :shrineMaiden: (:friend:), or :popIdol: (:keyPerson:) with the :richStudent: :murder: Incident on the 3rd Day.\n\n3rd Day\n\nFor the third Loop, try to kill the :shrineMaiden:.\n\nDay 1\n- Shrine Maiden: :paranoia:+1\n- :richStudent:: Move up or down\n- Pop Idol: :intrigue:+1\n\nIf the :richStudent: safely reaches the Shrine and is able to join the :shrineMaiden:, use her mandatory :conspiracyTheorist: ability to place :paranoia: on the :shrineMaiden:.\n\nDay 2\n- Shrine Maiden: :paranoia:+1\n- :boyStudent: Move left or right (avoid going to the Shrine)\n- Pop Idol: :intrigue:+1\n\nIf everything goes as planned up until now, the :shrineMaiden: (:friend:) will kill herself, so this Loop will end. If that doesn't work, try killing the :shrineMaiden: or :popIdol: on the third Day with :richStudent:'s :murder: Incident (if you miss this, too, you'll be in a lot of trouble lol).\n\n4th Loop\n\nIf you have been following the Hints, all Culprits and Roles should be revealed by now. Therefore, in the final 4th Loop, it'll be a battle of wits between the Mastermind and the Protagonists.\n\nGood luck"
    },
    {
        "title": "I'm Dying to Cause a Tragedy",
        "creator": "RooP",
        "source": "http://asuwa.mistysky.net/rooper/rooper.cgi",
        "difficultySets": [
            {
                "numberOfLoops": 4,
                "difficulty": 0
            }
        ],
        "tragedySet": "basicTragedy",
        "mainPlot": [
            "murderPlan"
        ],
        "subPlots": [
            "hiddenFreak",
            "threadsFate"
        ],
        "daysPerLoop": 5,
        "cast": {
            "doctor": "keyPerson",
            "informer": "killer",
            "officeWorker": "brain",
            "patient": "serialKiller",
            "richStudent": "friend",
            "boyStudent": "person",
            "girlStudent": "person",
            "shrineMaiden": "person",
            "policeOfficer": "person"
        },
        "incidents": [
            {
                "day": 2,
                "incident": "murder",
                "culprit": "shrineMaiden"
            },
            {
                "day": 3,
                "incident": "suicide",
                "culprit": "doctor"
            },
            {
                "day": 4,
                "incident": "increasingUnease",
                "culprit": "richStudent"
            },
            {
                "day": 5,
                "incident": "hospitalIncident",
                "culprit": "officeWorker"
            }
        ],
        "description": "Doctor I have no choice but to kill you\nIf that ojou-sama has such a tofu mentality, she has no choice but to die\nWho said that office worker is merely an ordinary person?",
        "specialRules": [
            "Table talk is not allowed. There is a Final Guess."
        ],
        "victory-conditions": "- Killing the :keyPerson:—:murder:, :suicide:, :hospitalIncident:, :serialKiller:, and :killer:.\n- Kill the Protagaonists—:hospitalIncident:, :killer:\n- Kill the :friend:—:murder:, :hospitalIncident:, :serialKiller:",
        "story": "A man who lost his daughter to a medical error swore revenge. \n\nHe gives him a knife, hiding it in a condolence gift. \"You're dying to kill someone, aren't you? Now's your chance.\"\n\nHe hands the woman a heavy suitcase. \"You can get anything you want as long you have enough money, right? Information, even a life.\"\n\nHe offers the cursed sword to the shrine. \"I want it exorcised within two days. It's dangerous if it gets rushed? That's not my concern, just get it done.\"\n\nHe looks into his eyes that are tormented by guilt. \"If you really feel bad about what you've done, can't you just apologize by dying?\"\n\nHe whispers to the young lady, who's surrounded by her clique. \"Hey, I have an interesting story for you. Care to listen?\"\n\nFinally, he ends it with his own hands. \"Let's go out with a bang.\"",
        "mastermindHints": "For the first Loop, you'll focus on the three students at the School, and you'll be able to avoid the Hospital. If there's no particular movement, you should be good.\n\nFor the second Loop, aim to kill the :richStudent: with a :murder: or the :serialKiller:.\n\nFor the third Loop, you'll be forced to choose between working behind the scenes with the Hospital, or working behind the scenes with the :doctor:. Furthermore, the :doctor: can be forced to choose between :paranoia: and :intrigue: until the third Day. The rest is flexible. The roles you need to hide are the :killer: and the :brain:. Use your ability after gathering a large number of people first."
    },
    {
        "title": "Escort Mission",
        "creator": "davidchaeB",
        "difficultySets": [
            {
                "numberOfLoops": 4,
                "difficulty": 0
            }
        ],
        "tragedySet": "basicTragedy",
        "mainPlot": [
            "murderPlan"
        ],
        "subPlots": [
            "paranoiaVirus",
            "threadsFate"
        ],
        "daysPerLoop": 6,
        "cast": {
            "richStudent": "keyPerson",
            "popIdol": "killer",
            "boss": [
                "brain",
                {
                    "Turf": "School"
                }
            ],
            "classRep": "conspiracyTheorist",
            "henchman": "person",
            "nurse": "person",
            "informer": "person",
            "mysteryBoy": "factor"
        },
        "incidents": [
            {
                "day": 2,
                "incident": "murder",
                "culprit": "classRep"
            },
            {
                "day": 4,
                "incident": "butterflyEffect",
                "culprit": "mysteryBoy"
            },
            {
                "day": 5,
                "incident": "farawayMurder",
                "culprit": "nurse"
            },
            {
                "day": 6,
                "incident": "missingPerson",
                "culprit": "richStudent"
            }
        ],
        "specialRules": [
            ""
        ],
        "victory-conditions": "",
        "story": "",
        "mastermindHints": "You have very strong power play. They can ruin it with Threads into Virus."
    }
]
```
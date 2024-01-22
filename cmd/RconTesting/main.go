package main

import (
	"fmt"
	"github.com/gorcon/rcon"
)

func main() {
	connection, err := rcon.Dial("54.80.201.244:25575", "<redacted>")
	if err != nil {
		panic(err)
	}
	defer connection.Close()

	executeCommand(connection, "ShowPlayers")
	executeCommand(connection, "Info")
	//executeCommand(connection, "Save")
	//executeCommand(connection, "Shutdown 5")
}

func executeCommand(connection *rcon.Conn, command string) {
	result, err := connection.Execute(command)
	if err != nil {
		fmt.Printf("%s error: %v\n", command, err)
	} else {
		fmt.Printf("%s result: %s\n", command, result)
	}
}

/*
OptionSettings=(Difficulty=None,
DayTimeSpeedRate=1.000000,
NightTimeSpeedRate=1.000000,
ExpRate=1.000000,
PalCaptureRate=1.000000,
PalSpawnNumRate=1.000000,
PalDamageRateAttack=1.000000,
PalDamageRateDefense=1.000000,
PlayerDamageRateAttack=1.000000,
PlayerDamageRateDefense=1.000000,
PlayerStomachDecreaceRate=1.000000,
PlayerStaminaDecreaceRate=1.000000,
PlayerAutoHPRegeneRate=1.000000,
PlayerAutoHpRegeneRateInSleep=1.000000,
PalStomachDecreaceRate=1.000000,
PalStaminaDecreaceRate=1.000000,
PalAutoHPRegeneRate=1.000000,
PalAutoHpRegeneRateInSleep=1.000000,
BuildObjectDamageRate=1.000000,
BuildObjectDeteriorationDamageRate=1.000000,
CollectionDropRate=1.000000,
CollectionObjectHpRate=1.000000,
CollectionObjectRespawnSpeedRate=1.000000,
EnemyDropItemRate=1.000000,
DeathPenalty=All,
bEnablePlayerToPlayerDamage=False,
bEnableFriendlyFire=False,
bEnableInvaderEnemy=True,
bActiveUNKO=False,
bEnableAimAssistPad=True,
bEnableAimAssistKeyboard=False,
DropItemMaxNum=3000,
DropItemMaxNum_UNKO=100,
BaseCampMaxNum=128,
BaseCampWorkerMaxNum=15,
DropItemAliveMaxHours=1.000000,
bAutoResetGuildNoOnlinePlayers=False,
AutoResetGuildTimeNoOnlinePlayers=72.000000,
GuildPlayerMaxNum=20,
PalEggDefaultHatchingTime=72.000000,
WorkSpeedRate=1.000000,
bIsMultiplay=False,
bIsPvP=False,
bCanPickupOtherGuildDeathPenaltyDrop=False,
bEnableNonLoginPenalty=True,
bEnableFastTravel=True,
bIsStartLocationSelectByMap=True,
bExistPlayerAfterLogout=False,
bEnableDefenseOtherGuildPlayer=False,
CoopPlayerMaxNum=4,
ServerPlayerMaxNum=32,
ServerName="Default Palworld Server",
ServerDescription="",
AdminPassword="SuperSecretAdminPassword",
ServerPassword="",
PublicPort=8211,
PublicIP="",
RCONEnabled=False,
RCONPort=25575,
Region="",
bUseAuth=True,
BanListURL="https://api.palworldgame.com/api/banlist.txt")
*/

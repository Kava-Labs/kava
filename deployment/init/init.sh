# Setup node and light client.
echo "This assumes you have wiped old data"

echo "copying kvd config to the default location ($HOME/.kvd)"
initData=./init-data
cp -vR $initData/.kvd/. $HOME/.kvd/ # copy contents of .kvd into .kvd, creating dir if necessary


echo "setup validator key from seed"
validatorPassword="$(cat $initData/validatorPassword)"
echo $validatorPassword
validatorBackupPhrase="$(cat $initData/validatorBackupPhrase)"
printf "$validatorPassword\n$validatorBackupPhrase\n" | kvcli keys add --recover validator


echo "setup user1 key from seed"
user1Password="$(cat ./init-data/user1Password)"
user1BackupPhrase="$(cat ./init-data/user1BackupPhrase)"
printf "$user1Password\n$user1BackupPhrase\n" | kvcli keys add --recover user1

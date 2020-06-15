#! /bin/bash

export USER=root
export HOME=/root
export ANDROID_HOME=$HOME/android-sdk
export ANDROID_SDK_ROOT=$ANDROID_HOME/
export PATH=${PATH}:$ANDROID_HOME/tools:$ANDROID_HOME/platform-tools:$ANDROID_HOME/emulators

source /root/.bashrc
/usr/bin/vncserver
export DISPLAY=:1
cd /root/android-sdk/emulator
./emulator -avd emulator -gpu swiftshader_indirect -no-snapshot-save & xterm -e "/usr/bin/appium -p 4444 --relaxed-security" & \
    sleep 20m; VMNAME=$(curl -H Metadata-Flavor:Google http://metadata/computeMetadata/v1/instance/hostname | cut -d. -f1); \
    ZONE=$(curl -H Metadata-Flavor:Google http://metadata/computeMetadata/v1/instance/zone | cut -d/ -f4); \
    gcloud compute instances delete $VMNAME --zone $ZONE --quiet

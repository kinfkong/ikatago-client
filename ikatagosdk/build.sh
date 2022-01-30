#!/bin/bash
gomobile bind --target=ios -o ../../ikatago-ios-sdk/ikatagosdk.framework .
rm -rf ~/gochess/react-native-ikatago-sdk/ios/ikatagosdk.framework
#gcp -RP ../../ikatago-ios-sdk/ikatagosdk.framework ~/gochess/react-native-ikatago-sdk/ios
gomobile bind --target=android -o ../../ikatago-android-sdk/ikatagosdk.aar .
#gcp -RP ../../ikatago-android-sdk/ikatagosdk.aar ~/gochess/react-native-ikatago-sdk/android/libs
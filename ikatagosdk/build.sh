#!/bin/bash
gomobile init
gomobile bind --target=ios -o ../../ikatago-ios-sdk/ikatagosdk.xcframework .
rm -rf ~/gochess/react-native-ikatago-sdk/ios/ikatagosdk.xcframework
gcp -RP ../../ikatago-ios-sdk/ikatagosdk.xcframework ~/gochess/react-native-ikatago-sdk/ios

gomobile bind --target=android -o ../../ikatago-android-sdk/ikatagosdk.aar .
gcp -RP ../../ikatago-android-sdk/ikatagosdk.aar ~/gochess/react-native-ikatago-sdk/android/libs
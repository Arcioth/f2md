{ pkgs ? import <nixpkgs> {} }:

let
  androidComposition = pkgs.androidenv.composeAndroidPackages {
    cmdLineToolsVersion = "20.0";
    buildToolsVersions = [ "35.0.0" ];
    platformVersions = [ "35" ];
    includeEmulator = false;
    includeSources = false;
    includeSystemImages = false;
    includeNDK = false;
    includeCmake = false;
  };
  androidSdk = androidComposition.androidsdk;
in
pkgs.mkShell {
  buildInputs = [
    androidSdk
    pkgs.jdk17
    pkgs.gradle
  ];

  ANDROID_HOME = "${androidSdk}/libexec/android-sdk";
  ANDROID_SDK_ROOT = "${androidSdk}/libexec/android-sdk";

  shellHook = ''
    export JAVA_HOME="${pkgs.jdk17}"
    export GRADLE_OPTS="-Dorg.gradle.project.android.aapt2FromMavenOverride=$(echo "$ANDROID_HOME/build-tools/"*"/aapt2")"
  '';
}

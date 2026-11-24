plugins {
    id("com.android.application")
    id("org.jetbrains.kotlin.android")
}

android {
    namespace = "com.arcioth.f2md"
    compileSdk = 35

    defaultConfig {
        applicationId = "com.arcioth.f2md"
        minSdk = 24
        targetSdk = 35
        versionCode = 4
        versionName = "0.4.0"

        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"
    }

    signingConfigs {
        getByName("debug") {
            enableV1Signing = true
            enableV2Signing = true
            enableV3Signing = true
        }
        create("release") {
            val keystoreFile = System.getenv("F2MD_KEYSTORE_PATH")?.let { file(it) }
                ?: rootProject.file("f2md-release.jks")
            if (keystoreFile.exists()) {
                storeFile = keystoreFile
                storePassword = System.getenv("F2MD_KEYSTORE_PASS") ?: "f2md-release-pass"
                keyAlias = System.getenv("F2MD_KEY_ALIAS") ?: "f2md"
                keyPassword = System.getenv("F2MD_KEY_PASS") ?: "f2md-release-pass"
            }
            enableV1Signing = true
            enableV2Signing = true
            enableV3Signing = true
        }
    }

    buildTypes {
        getByName("debug") {
            signingConfig = signingConfigs.getByName("debug")
        }
        release {
            isMinifyEnabled = false
            val relConfig = signingConfigs.getByName("release")
            signingConfig = if (relConfig.storeFile != null) relConfig else signingConfigs.getByName("debug")
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
        }
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    kotlinOptions {
        jvmTarget = "17"
    }

    testOptions {
        unitTests.isReturnDefaultValues = true
    }
}

dependencies {
    implementation("androidx.appcompat:appcompat:1.7.0")
    implementation("androidx.core:core-ktx:1.15.0")
    testImplementation("junit:junit:4.13.2")
}

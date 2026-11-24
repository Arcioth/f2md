# Keep JavaScript bridge methods exposed to WebView
-keepclassmembers class com.arcioth.f2md.MainActivity$F2mdBridge {
    @android.webkit.JavascriptInterface <methods>;
}

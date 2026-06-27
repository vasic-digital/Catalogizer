# google-services.json

This file contains the real Firebase credentials for production builds.
It is gitignored — obtain it from the Firebase Console or the project's
`.firebase/android-config.json` source-of-truth.

## Regeneration

```bash
firebase --project catalogizer-7a3f1 apps:sdkconfig android 1:881377664729:android:751a0d0e2d873db47768c8 > app/google-services.json
```

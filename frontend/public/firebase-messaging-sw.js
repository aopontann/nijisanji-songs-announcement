self.addEventListener("notificationclick", (event) => {
  console.log("event.notification:", event.notification);
  try {
    event.notification.close();
    clients.openWindow(event.notification.data?.url ?? "/");
  } catch (e) {
    // デバッグ用なので本番では消してもよいです
    console.error(e);
  }
});

importScripts(
  "https://www.gstatic.com/firebasejs/9.2.0/firebase-app-compat.js",
);
importScripts(
  "https://www.gstatic.com/firebasejs/9.2.0/firebase-messaging-compat.js",
);

// Initialize the Firebase app in the service worker by passing in
// your app's Firebase config object.
// https://firebase.google.com/docs/web/setup#config-object
firebase.initializeApp({
  apiKey: "AIzaSyBzGkZMCgRMfLa5MOPMDmycQT_Jb3wTQp8",
  authDomain: "niji-tuu.firebaseapp.com",
  projectId: "niji-tuu",
  storageBucket: "niji-tuu.appspot.com",
  messagingSenderId: "243582453217",
  appId: "1:243582453217:web:3c716c9d91edc5a1037ea0",
});

// Retrieve an instance of Firebase Messaging so that it can handle background
// messages.
const messaging = firebase.messaging();

// If you would like to customize notifications that are received in the
// background (Web app is closed or not in browser focus) then you should
// implement this optional method.
// Keep in mind that FCM will still show notification messages automatically
// and you should use data messages for custom notifications.
// For more info see:
// https://firebase.google.com/docs/cloud-messaging/concept-options
messaging.onBackgroundMessage(function (payload) {
  console.log(
    "[firebase-messaging-sw.js] Received background message ",
    payload,
  );

  const notificationTitle = payload.data.title;
  const notificationOptions = {
    body: payload.data.body,
    icon: payload.data.icon,
  };

  if (payload.data.type === "keyword") {
    keywordNotification(payload);
    return;
  }

  self.registration.showNotification(notificationTitle, notificationOptions);
});

const keywordNotification = (payload) => {
  const notificationTitle = payload.data.title;
  const notificationOptions = {
    body: payload.data.body,
    icon: payload.data.icon,
  };

  const openRequest = self.indexedDB.open("niji-tuu", 1);
  openRequest.onsuccess = function () {
    console.log("onsuccess");
    const db = openRequest.result;
    const transaction = db.transaction("keyword", "readonly");
    const objectStore = transaction.objectStore("keyword");
    const request = objectStore.get("1");

    request.onerror = (event) => {
      console.log("error:", request);
    };
    request.onsuccess = (event) => {
      // request.result に対して行う処理!
      console.log("text:", request.result.text);
      const reg = new RegExp(request.result.text);
      if (reg.test(payload.data.body)) {
        self.registration.showNotification(
          notificationTitle,
          notificationOptions,
        );
      }
    };
  };
};

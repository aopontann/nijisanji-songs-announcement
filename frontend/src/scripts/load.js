import { initializeApp } from "firebase/app";
import { fcmToken, firebaseConfig, vapidKey } from "./main";
import { getMessaging, getToken } from "firebase/messaging";

const songEle = document.getElementById("checkbox-song");
const infoEle = document.getElementById("checkbox-info");

initializeApp(firebaseConfig);

window.onload = async () => {
  console.log("load");

  // 既に購買済みか
  if (Notification.permission !== "granted") {
    return
  }

  const messaging = getMessaging();
  const currentToken = await getToken(messaging, { vapidKey });

  // APIサーバーから購買情報を取得　歌ってみた動画を通知する許可をしているか...
  const res = await fcmToken("GET", currentToken);
  if (res.status == 204) {
    console.log("no content");
    return
  }
  if (!res.ok) {
    window.alert("購買情報の取得に失敗しました。");
    return;
  }

  const data = await res.json();
  console.log("data", data);

  if (data == null) {
    return;
  }

  // 最新の購買情報に応じて要素を変更
  songEle.checked = data.song;
  window.localStorage.setItem("checkbox-song", data.song)
  infoEle.checked = data.info;
  window.localStorage.setItem("checkbox-info", data.info);
};

import { initializeApp } from "firebase/app";
import { firebaseConfig } from "./main";
import { fcmToken } from "./main";

const songEle = document.getElementById("checkbox-song");
const keywordEle = document.getElementById("checkbox-keyword");

initializeApp(firebaseConfig);

window.onload = async () => {
  console.log("load");

  // ブラウザーが通知に対応しているか調べる
  if (!("Notification" in window)) {
    window.alert("このブラウザーはデスクトップ通知には対応していません。");
    return;
  }

  // 既に購買済みか
  const currentToken = window.localStorage.getItem("fcm-token");
  if (currentToken == null) {
    return;
  }

  // APIサーバーから購買情報を取得　歌ってみた動画を通知する許可をしているか...
  const res = await fcmToken("GET", currentToken);
  if (!res.ok) {
    window.alert("購買情報の取得に失敗しました。");
    return;
  }

  const data = await res.json();
  console.log("data", data);

  if (data == null) {
    window.localStorage.clear("fcm-token");
    return;
  }

  // 最新の購買情報に応じて要素を変更
  songEle.checked = data.song;
  window.localStorage.setItem("checkbox-song", data.song)
  keywordEle.checked = data.keyword;
  window.localStorage.setItem("checkbox-keyword", data.keyword);
};

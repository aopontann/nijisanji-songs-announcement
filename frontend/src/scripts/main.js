import { initializeApp } from 'firebase/app';
import { deleteToken, getMessaging, getToken } from 'firebase/messaging';
import { firebaseConfig, vapidKey } from './config';

initializeApp(firebaseConfig);

const messaging = getMessaging();

const songEle = document.getElementById('checkbox-song')
const keywordEle = document.getElementById('checkbox-keyword')
const keywordTextEle = document.getElementById('keyword-text')

// if (window.navigator.serviceWorker !== undefined) {
//   window.navigator.serviceWorker.register('/firebase-messaging-sw.js');
// }

window.onload = async () => {
  console.log("page is fully loaded");

  // ブラウザーが通知に対応しているか調べる
  if (!("Notification" in window)) {
    window.alert("このブラウザーはデスクトップ通知には対応していません。");
    return
  }

  // 通知解除ボタン
  document.getElementById('delete').addEventListener("click", async () => {
    console.log("DELETE")
    document.getElementById("delete-spinner").style.display = "block"
    document.getElementById("submit-text").value = "通知解除中..."

    // FCMトークン削除
    const deleted = await deleteToken(messaging)
    if (!deleted) {
      window.alert('トークンの削除に失敗しました')
      console.log("トークンの削除に失敗しました")
    }

    window.localStorage.clear("fcm-token")

    const res = await fcmToken('DELETE', currentToken)
    if (!res.ok) {
      window.alert('トークンの削除に失敗しました')
      console.log("トークンの削除に失敗しました")
    }
    else {
      // document.getElementById("toast-success").style.display = "block"
      document.getElementById("toast-success").style.visibility = "visible"
    }

    document.getElementById("delete-spinner").style.display = "none"
    document.getElementById("submit-text").value = "通知解除"
    songEle.checked = false
    keywordEle.checked = false
  })

  // ---------------------------------------------------------------------------------------- //

  // 既に購買済みか
  const currentToken = window.localStorage.getItem("fcm-token")
  if (currentToken == null) {
    return
  }
  
  // キーワードを書き込む
  const openRequest = window.indexedDB.open("niji-tuu", 1);
  openRequest.onsuccess = function () {
    console.log("onsuccess")
    const db = openRequest.result;
    const transaction = db.transaction("keyword", "readonly");
    const objectStore = transaction.objectStore("keyword");
    const request = objectStore.get("1");

    request.onerror = (event) => {
      console.log("error:", request)
    };
    request.onsuccess = (event) => {
      // request.result に対して行う処理!
      console.log("text:", request.result.text)
      keywordTextEle.value = request.result.text
    };
  };

  document.getElementById("overlay").style.display = "block"

  // APIサーバーから購買情報を取得　歌ってみた動画を通知する許可をしているか...
  const res = await fcmToken('GET', currentToken)
  if (!res.ok) {
    window.alert("購買情報の取得に失敗しました。")
    document.getElementById("overlay").style.display = "none"
    return
  }

  const data = await res.json()
  console.log("data", data)

  if (data == null) {
    document.getElementById("overlay").style.display = "none"
    keywordTextEle.value = window.localStorage.getItem("word")
    window.localStorage.clear("fcm-token")
    return
  }

  // 最新の購買情報に応じて要素を変更
  if (data.song == 1) {
    songEle.checked = true
  }
  if (data.keyword == 1) {
    keywordEle.checked = true
  }

  document.getElementById("overlay").style.display = "none"
};


async function fcmToken(method, token) {
  console.log('token:', token)
  return await fetch("/api/subscription", {
    method,
    cache: "no-cache",
    headers: { "Token": JSON.stringify(token) },
  });
}

import { initializeApp } from 'firebase/app';
import { deleteToken, getMessaging, getToken } from 'firebase/messaging';
import { firebaseConfig, vapidKey } from './config';

initializeApp(firebaseConfig);

const messaging = getMessaging();

const songEle = document.getElementById('song')
const keywordEle = document.getElementById('keyword')
const keywordTextEle = document.getElementById('keyword-text')

if (window.navigator.serviceWorker !== undefined) {
  window.navigator.serviceWorker.register('/firebase-messaging-sw.js');
}

window.onload = async () => {
  console.log("page is fully loaded");

  // ブラウザーが通知に対応しているか調べる
  if (!("Notification" in window)) {
    window.alert("このブラウザーはデスクトップ通知には対応していません。");
    return
  }

  // ------------------------------- ボタン クリックイベント処理 ------------------------------- //
  // 保存ボタン
  // document.getElementById('submit').addEventListener('click', async () => {
  //   console.log("POST")
  //   document.getElementById("submit-spinner").style.display = "block"
  //   document.getElementById("submit-text").value = "通知登録中..."

  //   console.log("generate token...")
  //   var currentToken = ""
  //   try {
  //     currentToken = await getToken(messaging, { vapidKey })
  //     console.log("generated token:", currentToken)
  //   } catch (error) {
  //     document.getElementById("submit-spinner").style.display = "none"
  //     document.getElementById("submit-text").value = "通知登録"
  //     // 通知権限がブロックされている場合
  //     if (Notification.permission === "denied") {
  //       window.alert('通知がブロックされています。')
  //       return
  //     }

  //     // 通知権限がブロックされていないが、ユーザーの許可を得れていない場合
  //     if (Notification.permission === "default") {
  //       window.alert('通知を許可してください。')
  //       return
  //     }

  //     // 通知権限がブロックされていないが、ユーザーの許可を得れていない場合
  //     if (Notification.permission === "granted") {
  //       window.alert('エラーが発生しました。再度通知登録をしてください。')
  //       return
  //     }
  //   }

  //   if (currentToken == "") {
  //     return
  //   }
  //   console.log("songEle", songEle)
  //   console.log("keywordTextEle", keywordTextEle)

  //   const ok = await saveToken(currentToken, { song: songEle.checked ? 1 : 0, word: keywordEle.checked && keywordTextEle.value || "" })
  //   if (!ok) {
  //     window.alert('トークンの登録に失敗しました')
  //     console.log("トークンの登録に失敗しました")
  //   }

  //   window.localStorage.setItem("fcm-token", currentToken)
  //   // document.getElementById("toast-success").style.display = "block"
  //   document.getElementById("toast-success").style.visibility = "visible"

  //   document.getElementById("submit-spinner").style.display = "none"
  //   document.getElementById("submit-text").value = "通知登録"
  // })
  
  // 通知解除ボタン
  document.getElementById('delete').addEventListener("click", async () => {
    console.log("DELETE")
    document.getElementById("delete-spinner").style.display = "block"
    document.getElementById("submit-text").value = "通知解除中..."

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
        console.log("text:", request.result)
      };
  };

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
  if (data.word != '') {
    keywordTextEle.value = data.word
    keywordEle.checked = true
  }
  if (data.word == '') {
    keywordTextEle.value = window.localStorage.getItem("word")
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

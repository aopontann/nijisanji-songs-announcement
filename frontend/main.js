import { initializeApp } from './node_modules/firebase/app';
import { deleteToken, getMessaging, getToken } from './node_modules/firebase/messaging';
import { firebaseConfig, vapidKey } from './config';

initializeApp(firebaseConfig);

const messaging = getMessaging();

const songEle = document.getElementById('song')
const keywordEle = document.getElementById('keyword')
const keywordTextEle = document.getElementById('keyword-text')

// if (window.navigator.serviceWorker !== undefined) {
//   window.navigator.serviceWorker.register('/firebase-messaging-sw.js');
// }

window.onload = async () => {
  console.log("page is fully loaded");

  // iPhone, iPad の場合
  if (isIos() && !isInStandaloneMode()) {
    document.getElementById('WarningDialog').showModal();
    return
  }

  // --------------- 利用規約・プライバシーポリシー 関連の処理 --------------- //
  document.getElementById("ok").addEventListener("click", () => {
    window.localStorage.setItem("readed-terms", "true")
    document.getElementById("favDialog").close()
  })

  if (window.localStorage.getItem("readed-terms") !== "true") {
    document.getElementById("favDialog").showModal();
  }
  // -------------------------------------------------------------------- //

  // ブラウザーが通知に対応しているか調べる
  if (!("Notification" in window)) {
    window.alert("このブラウザーはデスクトップ通知には対応していません。");
    return
  }

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

  // ------------------------------- ボタン クリックイベント処理 ------------------------------- //
  // 保存ボタン
  document.getElementById('submit').addEventListener('click', async () => {
    console.log("POST")
    document.getElementById("overlay").style.display = "block"

    console.log("generate token...")
    var currentToken = ""
    try {
      currentToken = await getToken(messaging, { vapidKey })
      console.log("generated token:", currentToken)
    } catch (error) {
      document.getElementById("overlay").style.display = "none"
      // 通知権限がブロックされている場合
      if (Notification.permission === "denied") {
        window.alert('通知がブロックされています')
        return
      }

      // 通知権限がブロックされていないが、ユーザーの許可を得れていない場合
      if (Notification.permission === "default") {
        window.alert('通知を許可してください')
        return
      }
    }

    if (currentToken == "") {
      return
    }

    const ok = await saveToken(currentToken, { song: songEle.checked ? 1 : 0, word: keywordEle.checked && keywordTextEle.value || "" })
    if (!ok) {
      window.alert('トークンの登録に失敗しました')
      console.log("トークンの登録に失敗しました")
    }

    window.localStorage.setItem("fcm-token", currentToken)

    document.getElementById("overlay").style.display = "none"
  })

  // 通知解除ボタン
  document.getElementById('delete').addEventListener("click", async () => {
    console.log("DELETE")
    document.getElementById("overlay").style.display = "block"

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

    document.getElementById("overlay").style.display = "none"
    songEle.checked = false
    keywordEle.checked = false
  })

  // キーワード
  keywordTextEle.addEventListener("change", (event) => {
    window.localStorage.setItem("word", event.target.value)
  })
  // ---------------------------------------------------------------------------------------- //
};

const isIos = () => {
  const userAgent = window.navigator.userAgent.toLowerCase();
  return /iphone|ipad|ipod/.test(userAgent);
};

// check if the device is in standalone mode
const isInStandaloneMode = () => {
  return window.matchMedia('(display-mode: standalone)').matches;
};

async function fcmToken(method, token) {
  console.log('token:', token)
  return await fetch("/api/subscription", {
    method,
    cache: "no-cache",
    headers: { "Token": JSON.stringify(token) },
  });
}

async function saveToken(token, data = {}) {
  const url = "/api/subscription"
  const response = await fetch(url, {
    method: "POST",
    cache: "no-cache",
    headers: {
      "Content-Type": "application/json",
      "Token": JSON.stringify(token),
    },
    body: JSON.stringify(data), // 本体のデータ型は "Content-Type" ヘッダーと一致させる必要があります
  });

  console.log(response.status)
  console.log(response.ok)
  return response.ok; // JSON のレスポンスをネイティブの JavaScript オブジェクトに解釈
}
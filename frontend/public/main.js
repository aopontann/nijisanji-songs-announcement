const permissionEle = document.getElementById('permission')
const msgEle = document.getElementById('msg')
const songEle = document.getElementById('song')
const keywordEle = document.getElementById('keyword')
const keywordTextEle = document.getElementById('keyword-text')
const submitEle = document.getElementById('submit')
const deleteEle = document.getElementById('delete')
const confEle = document.getElementById('subscription-conf')
const loadingEle = document.getElementById('loading')
const dialogEle = document.getElementById("favDialog");
const okEle = document.getElementById("ok");

if (window.navigator.serviceWorker !== undefined) {
  window.navigator.serviceWorker.register('/serviceworker.js');
}

// ローディング画面の実装

window.onload = async () => {
  console.log("page is fully loaded");

  okEle.addEventListener("click", () => {
    window.localStorage.setItem("readed-terms", "true")
    dialogEle.close()
  })

  if (window.localStorage.getItem("readed-terms") !== "true") {
    dialogEle.showModal();
  }

  // iPhone, iPad の場合
  if (isIos() && !isInStandaloneMode()) {
    msgEle.innerText = "iPhone, iPad, iPodをお使いの場合、「共有」→「ホーム画面に追加」して、追加されたアイコンから起動してください"
    return
  }

  // ブラウザーが通知に対応しているか調べる
  if (!("Notification" in window)) {
    window.alert("このブラウザーはデスクトップ通知には対応していません。");
    return
  }

  // 通知の許可が既に得られている場合
  if (Notification.permission === "granted") {
    // 何もしない (o.o)
  }

  // 通知権限がブロックされている場合
  if (Notification.permission === "denied") {
    window.alert('通知がブロックされています')
    return
  }

  // 通知権限がブロックされていないが、ユーザーの許可を得れていない場合
  if (Notification.permission === "default") {
    permissionEle.style.display = "block"
    msgEle.innerText = "↑の通知許可ボタンを押して、通知を許可してください"

    permissionEle.addEventListener("click", async () => {
      // 通知許可を求める
      await Notification.requestPermission()
      window.location.reload()
    })
    return
  }

  msgEle.style.display = "none"
  document.getElementById("overlay").style.display = "block"

  const registration = await navigator.serviceWorker.ready

  submitEle.addEventListener('click', async () => {
    console.log("POST")

    document.getElementById("overlay").style.display = "block"

    const subscription = await registration.pushManager.getSubscription() || await registration.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: "BCSvj0H4g72CXuyK_CUy2oygQyRXDyX_BaR2ACtfmEYm2jLj-qCymSnDhfp7acuBISkKxj_UC1TKd6eOPcfr27w",
    })

    const isOK = await fetchData("POST", subscription.toJSON(), { song: songEle.checked ? 1 : 0, word: keywordEle.checked && keywordTextEle.value || "" })
    if (!isOK) {
      window.alert('トークンの登録に失敗しました')
      console.log("トークンの登録に失敗しました")
    }
    document.getElementById("overlay").style.display = "none"
  })

  deleteEle.addEventListener("click", async () => {
    console.log("DELETE")

    const subscription = await registration.pushManager.getSubscription() || await registration.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: "BCSvj0H4g72CXuyK_CUy2oygQyRXDyX_BaR2ACtfmEYm2jLj-qCymSnDhfp7acuBISkKxj_UC1TKd6eOPcfr27w",
    })

    document.getElementById("overlay").style.display = "block"
    const isOK = await fetchData("DELETE", subscription.toJSON())
    if (!isOK) {
      window.alert('トークンの削除に失敗しました')
      console.log("トークンの削除に失敗しました")
      return
    }
    document.getElementById("overlay").style.display = "none"
    songEle.checked = false
    keywordEle.checked = false
  })

  keywordTextEle.addEventListener("change", (event) => {
    window.localStorage.setItem("word", event.target.value)
  })

  // 既に購買済みか
  const subscription = await registration.pushManager.getSubscription()
  console.log("subscription:", subscription)
  if (subscription == null) {
    confEle.style.display = 'block'
    document.getElementById("overlay").style.display = "none"
    return
  }

  // APIサーバーから購買情報を取得　歌ってみた動画を通知する許可をしているか...
  const res = await isSubscription(subscription.toJSON())
  if (res.status == 404) {
    confEle.style.display = 'block'
    document.getElementById("overlay").style.display = "none"
    keywordTextEle.value = window.localStorage.getItem("word")
    return
  } else if (!res.ok) {
    window.alert("購買情報の取得に失敗しました。")
    return
  }

  // 最新の購買情報に応じて要素を変更
  const data = await res.json()
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

  confEle.style.display = 'block'
  document.getElementById("overlay").style.display = "none"
};

const isIos = () => {
  const userAgent = window.navigator.userAgent.toLowerCase();
  return /iphone|ipad|ipod/.test(userAgent);
};

// check if the device is in standalone mode
const isInStandaloneMode = () => {
  return window.matchMedia('(display-mode: standalone)').matches;
};

async function isSubscription(token) {
  const url = "/api/subscription"
  const response = await fetch(url, {
    method: "GET",
    cache: "no-cache",
    headers: {
      "Token": JSON.stringify(token),
    },
  });

  return response
}

async function fetchData(method = "", token, data = {}) {
  const url = "/api/subscription"
  const response = await fetch(url, {
    method,
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
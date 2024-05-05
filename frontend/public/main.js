const msgEle = document.getElementById('msg')
const songEle = document.getElementById('song')
const keywordEle = document.getElementById('keyword')
const keywordTextEle = document.getElementById('keyword-text')
const submitEle = document.getElementById('submit')
const deleteEle = document.getElementById('delete')
const confEle = document.getElementById('subscription-conf')
const loadingEle = document.getElementById('loading')

if (window.navigator.serviceWorker !== undefined) {
  window.navigator.serviceWorker.register('/serviceworker.js');
}

window.onload = async () => {
  console.log("page is fully loaded");

  // iPhone, iPad の場合
  if (isIos() && !isInStandaloneMode()) {
    window.alert("iPhone, iPad, iPodをお使いの場合、「共有」→「ホーム画面に追加」して、追加されたアイコンから起動してください")
    msgEle.innerText = "iPhone, iPad, iPodをお使いの場合、「共有」→「ホーム画面に追加」して、追加されたアイコンから起動してください"
    return
  } 

  // サービスワーカー or WebPush に対応していない場合
  if (!await checkIsWebPushSupported()) {
    window.alert("お使いのブラウザでは、サービスワーカー or WebPush に対応していません")
    msgEle.innerText = "お使いのブラウザでは、サービスワーカー or WebPush に対応していません"
  return
  }

  // 通知許可を求める
  if (!await IsNotificationAllowed()) {
    return
  }

  const registration = await navigator.serviceWorker.ready

  // Pushマネージャーから登録済み購買情報を取得
  // 取得できない場合、Pushサーバーに購買情報を登録
  const subscription = await registration.pushManager.getSubscription() || await registration.pushManager.subscribe({
    userVisibleOnly: true,
    applicationServerKey: "BCSvj0H4g72CXuyK_CUy2oygQyRXDyX_BaR2ACtfmEYm2jLj-qCymSnDhfp7acuBISkKxj_UC1TKd6eOPcfr27w",
  })

  submitEle.addEventListener('click', async () => {
    console.log("POST")
    loadingEle.style.display = 'block'
    const isOK = await fetchData("POST", subscription.toJSON(), { song: songEle.checked ? 1 : 0, word: keywordEle.checked && keywordTextEle.value || "" })
    if (!isOK) {
      window.alert('トークンの登録に失敗しました')
      console.log("トークンの登録に失敗しました")
    }
    loadingEle.style.display = 'none'
  })

  deleteEle.addEventListener("click", async () => {
    console.log("DELETE")
    loadingEle.style.display = 'block'
    const isOK = await fetchData("DELETE", subscription.toJSON())
    if (!isOK) {
      window.alert('トークンの削除に失敗しました')
      console.log("トークンの削除に失敗しました")
      return
    }
    loadingEle.style.display = 'none'
    songEle.checked = false
    keywordEle.checked = false
  })

  keywordTextEle.addEventListener("change", (event) => {
    window.localStorage.setItem("word", event.target.value)
  })

  // APIサーバーから購買情報を取得　歌ってみた動画を通知する許可をしているか...
  const res = await isSubscription(subscription.toJSON())
  if (res.status == 404) {
    confEle.style.display = 'block'
    loadingEle.style.display = 'none'
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
  loadingEle.style.display = 'none'
};

const isIos = () => {
  const userAgent = window.navigator.userAgent.toLowerCase();
  return /iphone|ipad|ipod/.test(userAgent);
};

// check if the device is in standalone mode
const isInStandaloneMode = () => {
  return ("standalone" in window.navigator && window.navigator.standalone);
};

const checkIsWebPushSupported = async () => {
  // グローバル空間にNotificationがあればNotification APIに対応しているとみなす
  if (!('Notification' in window)) {
    return false;
  }
  // グローバル変数navigatorにserviceWorkerプロパティがあればサービスワーカーに対応しているとみなす
  if (!('serviceWorker' in navigator)) {
    return false;
  }
  try {
    const sw = await navigator.serviceWorker.ready;
    // 利用可能になったサービスワーカーがpushManagerプロパティがあればPush APIに対応しているとみなす
    if (!('pushManager' in sw)) {
      return false;
    }
    return true;
  } catch (error) {
    return false;
  }
}

const IsNotificationAllowed = async () => {
  const permission = await window.Notification.requestPermission()
  if (permission === 'denied') {
    window.alert('通知がブロックされています')
    msgEle.innerText = "通知がブロックされています"
    return false
  }
  else if (permission === 'granted') {
    const registration = await navigator.serviceWorker.ready
    console.log("サービスワーカーがアクティブ:", registration.active);
    return true
  }
  else if (permission === 'default') {
    window.alert('通知を許可してください')
    msgEle.innerText = "通知を許可してください"
    return false
  }
  else {
    console.log('Unable to get permission to notify.');
    msgEle.innerText = "Unable to get permission to notify."
    return false
  }
}

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

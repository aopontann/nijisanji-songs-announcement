const checkbox = document.getElementById('subscribe')
const confEle = document.getElementById('subscription-conf')
const loadingEle = document.getElementById('loading')

if (window.navigator.serviceWorker !== undefined) {
  window.navigator.serviceWorker.register('/serviceworker.js');
}

window.onload = async() => {
  console.log("page is fully loaded");

  checkbox.addEventListener('change', async(event) => {
    console.log("AAA")
    checkbox.disabled = true
    if (event?.currentTarget?.checked) {
      await subscribe()
    } else {
      await unSubscribe()
    }
    checkbox.disabled = false
  })

  if(!await checkIsWebPushSupported()) {
    window.alert('「共有」→「ホーム画面に追加」して、追加されたアイコンから起動してください')
    return
  }
  try {
    const registration = await navigator.serviceWorker.ready
    const subscription = await registration.pushManager.getSubscription()
    if (!subscription) {
      confEle.style.display = 'block'
      loadingEle.style.display = 'none'
      return
    }
    const isOK = await postData('/api/subscription/check', subscription.toJSON())
    console.log('isOK:', isOK)
    if (isOK) {
      checkbox.checked = true
    } else {
      checkbox.checked = false
    }
  } catch (error) {
      window.alert('予期せぬエラーが発生しました')
      checkbox.checked = false
  }
  confEle.style.display = 'block'
  loadingEle.style.display = 'none'
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
    return false
  }
  else if (permission === 'granted') {
    const registration = await navigator.serviceWorker.ready
    console.log("サービスワーカーがアクティブ:", registration.active);
    return true
  }
  else if (permission === 'default') {
    window.alert('通知を許可してください')
    return false
  }
  else {
    console.log('Unable to get permission to notify.');
    return false
  }
}

async function subscribe() {
  if (!(await IsNotificationAllowed())) {
    checkbox.checked = false
    return
  }
  console.log("subscribe")
  const registration = await navigator.serviceWorker.ready
  try {
    const currentLocalSubscription = await registration.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: "BCSvj0H4g72CXuyK_CUy2oygQyRXDyX_BaR2ACtfmEYm2jLj-qCymSnDhfp7acuBISkKxj_UC1TKd6eOPcfr27w",
    })
    console.log('currentLocalSubscription:', currentLocalSubscription.toJSON())
    const isOK = await postData('/api/subscription', currentLocalSubscription.toJSON())
    if (isOK) {
      checkbox.checked = true
    }
    else {
      checkbox.checked = false
      console.log("トークンの登録に失敗しました")
    }
  } catch (error) {
    console.log(error)
    checkbox.checked = false
  }
}

async function unSubscribe() {
  console.log("unSubscribe")
  const registration = await navigator.serviceWorker.ready
  const subscription = await registration.pushManager.getSubscription()
  try {
    if (!(await subscription?.unsubscribe())) {
      console.log("通知の解除に失敗しました")
      checkbox.checked = true
      return
    }
    const isOK = await deleteData('/api/subscription', subscription?.toJSON())
    if (isOK) {
      checkbox.checked = false
    }
    else {
      checkbox.checked = true
      console.log("トークンの削除に失敗しました")
    }
  } catch (error) {
    console.log(error)
  }
}

async function postData(url = "", data = {}) {
  // 既定のオプションには * が付いています
  const response = await fetch(url, {
    method: "POST", // *GET, POST, PUT, DELETE, etc.
    mode: "cors", // no-cors, *cors, same-origin
    cache: "no-cache", // *default, no-cache, reload, force-cache, only-if-cached
    credentials: "same-origin", // include, *same-origin, omit
    headers: {
      "Content-Type": "application/json",
      // 'Content-Type': 'application/x-www-form-urlencoded',
    },
    redirect: "follow", // manual, *follow, error
    referrerPolicy: "no-referrer", // no-referrer, *no-referrer-when-downgrade, origin, origin-when-cross-origin, same-origin, strict-origin, strict-origin-when-cross-origin, unsafe-url
    body: JSON.stringify(data), // 本体のデータ型は "Content-Type" ヘッダーと一致させる必要があります
  });
  console.log(response.status)
  console.log(response.ok)
  return response.ok; // JSON のレスポンスをネイティブの JavaScript オブジェクトに解釈
}

async function deleteData(url = "", data = {}) {
  // 既定のオプションには * が付いています
  const response = await fetch(url, {
    method: "DELETE", // *GET, POST, PUT, DELETE, etc.
    mode: "cors", // no-cors, *cors, same-origin
    cache: "no-cache", // *default, no-cache, reload, force-cache, only-if-cached
    credentials: "same-origin", // include, *same-origin, omit
    headers: {
      "Content-Type": "application/json",
      // 'Content-Type': 'application/x-www-form-urlencoded',
    },
    redirect: "follow", // manual, *follow, error
    referrerPolicy: "no-referrer", // no-referrer, *no-referrer-when-downgrade, origin, origin-when-cross-origin, same-origin, strict-origin, strict-origin-when-cross-origin, unsafe-url
    body: JSON.stringify(data), // 本体のデータ型は "Content-Type" ヘッダーと一致させる必要があります
  });
  return response.ok; // JSON のレスポンスをネイティブの JavaScript オブジェクトに解釈
}
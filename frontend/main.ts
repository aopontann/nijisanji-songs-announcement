const checkbox = document.getElementById('subscribe')

if (window.navigator.serviceWorker !== undefined) {
  window.navigator.serviceWorker.register('/serviceworker.js');
}

window.onload = async() => {
  console.log("page is fully loaded");
  try {
    const registration = await navigator.serviceWorker.ready
    const subscription = await registration.pushManager.getSubscription()
    const isOK = await postData('/is-subscription', subscription?.toJSON())
    if (isOK) {
      window.localStorage.setItem('isSubscribe', '1');
      checkbox.checked = true
    } else {
      window.localStorage.setItem('isSubscribe', '0');
      checkbox.checked = false
    }
  } catch (error) {
    window.localStorage.setItem('isSubscribe', '0');
      checkbox.checked = false
  }
};

const isSupported = async () => {
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
  else {
    console.log('Unable to get permission to notify.');
    return false
  }
}

async function subscribe() {
  if (!(await isSupported())) {
    checkbox.checked = false
  }
  console.log("subscribe")
  const registration = await navigator.serviceWorker.ready
  try {
    const currentLocalSubscription = await registration.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: "BCSvj0H4g72CXuyK_CUy2oygQyRXDyX_BaR2ACtfmEYm2jLj-qCymSnDhfp7acuBISkKxj_UC1TKd6eOPcfr27w",
    })
    console.log('currentLocalSubscription:', currentLocalSubscription.toJSON())
    const isOK = await postData('/token', currentLocalSubscription.toJSON())
    if (isOK) {
      window.localStorage.setItem('isSubscribe', '1');
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
      return
    }
    const isOK = await deleteData('/token', subscription?.toJSON())
    if (isOK) {
      window.localStorage.setItem('isSubscribe', '0');
    }
    else {
      checkbox.checked = false
      console.log("トークンの削除に失敗しました")
    }
  } catch (error) {
    console.log(error)
  }
}

checkbox?.addEventListener('change', (event) => {
  if (event?.currentTarget?.checked) {
    // alert('checked');
    subscribe()
  } else {
    // alert('not checked');
    unSubscribe()
  }
})


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
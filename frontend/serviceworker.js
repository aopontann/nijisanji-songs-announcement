self.addEventListener('push', (event) => {
    try {
        if (
            self.Notification == null ||
            self.Notification.permission !== 'granted'
        ) {
            console.debug('notification is disabled.');
            return;
        }
        console.log('event:', event)

        const payload = event.data?.json() ?? null;
        const title = payload?.title ?? 'プッシュ通知で表示されるタイトルのデフォルト値';
        const tag = payload?.tag ?? '';
        const body = payload?.body ?? '';
        const icon = payload?.icon ?? 'プッシュ通知で表示させたいアイコン画像URLのデフォルト値';
        const data = payload?.data ?? null;

        self.registration.showNotification(title, {
            body,
            tag,
            icon,
            data,
        });
    } catch (e) {
        // デバッグ用なので本番では消してもよいです
        console.error(e);
    }
});

self.addEventListener('notificationclick', (event) => {
    try {
        event.notification.close();
        clients.openWindow(event.notification.data?.url ?? '/');
    } catch (e) {
        // デバッグ用なので本番では消してもよいです
        console.error(e);
    }
});
export async function onRequestPost(context) {
    const json = await context.request.json()
    const text = JSON.stringify(json)
    const now = new Date().toISOString()

    try {
        const res = await context.env.MY_DB.prepare('UPDATE users SET time = ?1 WHERE token = ?2').bind(now, text).run();
        console.log(res)
        if (res.success)
            return new Response("OK")
        else
            return new Response("NG", { status: 400 })
    } catch (error) {
        return new Response("NG", { status: 500 })
    }
}
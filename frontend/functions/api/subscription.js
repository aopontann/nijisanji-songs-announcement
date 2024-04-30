/**
 * /api/subscription
 */
export async function onRequestPost(context) {
    const json = await context.request.json()
    const text = JSON.stringify(json)
    
    try {
        const res = await context.env.MY_DB.prepare('INSERT INTO users (token, song, word, time) VALUES (?1, ?2, ?3, ?4)').bind(text, 1, '', '').run()
        console.log(res)
        if (res.success)
            return new Response("OK")
        else
            return new Response("NG", null, 400)
    } catch (error) {
        return new Response("NG", null, 500)
    }
}

export async function onRequestDelete(context) {
    const json = await context.request.json()
    const text = JSON.stringify(json)

    try {
        const res = await context.env.MY_DB.prepare('DELETE FROM users WHERE token ?1').bind(text).run()
        console.log(res)
        if (res.success)
            return new Response("OK")
        else
            return new Response("NG", null, 400)
    } catch (error) {
        return new Response("NG", null, 500)
    }
}

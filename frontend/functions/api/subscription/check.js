export async function onRequestPost(context) {
    const json = await context.request.json()
    const text = JSON.stringify(json)

    try {
        const total = await context.env.MY_DB.prepare('SELECT COUNT(*) AS total FROM users WHERE token = ?1').bind(text).first('total');
        console.log(total); // 50
        console.log(typeof (total)); // 50
        if (total === 0)
            return new Response("NG", null, 404)
        else
            return new Response('OK')
    } catch (error) {
        return new Response("NG", null, 500)
    }
}
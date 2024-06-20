/**
 * /api/subscription
 */
export async function onRequestGet(context) {
  console.log(context.request.headers);
  const token = context.request.headers.get("Token");

  try {
    const stmt = context.env.MY_DB.prepare(
      "SELECT song, keyword FROM users WHERE token = ?1 LIMIT 1",
    ).bind(token);
    const res = await stmt.first();
    console.log(res);

    if (res == null) {
      return new Response(null);
    } else {
      return new Response(JSON.stringify(res), {
        headers: new Headers({
          "Content-Type": "application/json",
        }),
      });
    }
  } catch (error) {
    return new Response("NG", { status: 500 });
  }
}

export async function onRequestPost(context) {
  const json = await context.request.json();
  console.log(context.request.headers);
  console.log("json:", json);
  const token = context.request.headers.get("Token");
  console.log("token:", token);

  try {
    const res = await context.env.MY_DB.prepare(
      "INSERT INTO users (token, song, keyword, keyword_text) VALUES (?1, ?2, ?3, ?4) ON CONFLICT(token) do update set song = ?2, keyword = ?3, keyword_text = ?4",
    )
      .bind(token, json.song, json.keyword, json.keyword_text)
      .run();
    console.log(res);
    if (res.success) return new Response("OK");
    else return new Response("NG", { status: 400 });
  } catch (error) {
    console.log(error);
    return new Response("NG", { status: 500 });
  }
}

export async function onRequestDelete(context) {
  console.log(context.request.headers);
  const token = context.request.headers.get("Token");
  console.log(token);

  try {
    const res = await context.env.MY_DB.prepare(
      "DELETE FROM users WHERE token = ?1",
    )
      .bind(token)
      .run();
    console.log(res);
    if (res.success) return new Response("OK");
    else return new Response("NG", { status: 400 });
  } catch (error) {
    console.error(error);
    return new Response("NG", { status: 500 });
  }
}

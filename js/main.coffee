hash = window.location.hash.substr(1)

fun_res = (result, item) ->
    parts = item.split('=')
    result[parts[0]] = parts[1]
    result

result = hash.split('&').reduce fun_res, {}

if result.access_token != undefined
    url = 'http://localhost:5454/access_token?access_token='+result.access_token
    window.location = url

import requests as r, hashlib as h
from urllib.parse import quote, quote_plus

def map_api_baidu(lat: float, lon: float, ak: str, sk: str) -> r.Response:
    path = "/reverse_geocoding/v3/?"
    params = f"location={lat},{lon}&coordtype=wgs84ll&output=json&ak={ak}"
    query_str = path + params
    encoded_str = quote(query_str, safe="/:=&?#+!$,;'@()*[]")
    raw_str = encoded_str + sk
    sn = h.md5(quote_plus(raw_str).encode('utf-8')).hexdigest()
    final_url = f"http://api.map.baidu.com{query_str}&sn={sn}"
    return r.get(final_url)


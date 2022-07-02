import * as http from "http"
import * as https from "https"
import * as cmd from "process"
import * as fs from "fs"
import { RequestOptions } from "https";
import { Transform } from "stream";

var args = cmd.argv.slice(2)

var stream = Transform

interface IngaiaResponse {
  location_street_number: Number,
  usage: string,
  rent_price: number,
  photos: {
    big: string
  }[],
  rent_average_price: number,
  municipal_property_tax: number,
  usage_type: string,
  property_reference: string,
  location_street_address: string,
  bathrooms: number,
  garages: number,
  bedroom_bath: number,
  sale_average_price: number,
  location_add_on_address: string,
  location_neighborhood: string,
  beds: number,
  area_useful: number,
  enterprise: string,
  location_state: string,
  total_garages: number,
  area_built: number,
  location_city: string,
  has_negotiation: boolean,
  condo_price: number,
  agency_name: string,
  sale_price: number,
  has_proposal: boolean,
  area: number,
  area_label: string,
}

var req: RequestOptions = {
  host: "listings.ingaia.com.br",
  path: `/listings?page_num=0&per_page=12&scope=Agency&property_reference=${args[0]}`,
  method: "GET",
  headers: {
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:99.0) Gecko/20100101 Firefox/99.0",
    "Accept": "application/json, text/plain, */*",
    "Accept-Language": "pt-BR,pt;q=0.8,en-US;q=0.5,en;q=0.3",
    "Accept-Encoding": "gzip, deflate, br",
    "Origin": "https://imob.valuegaia.com.br",
    "Sec-Fetch-Dest": "empty",
    "Sec-Fetch-Mode": "no-cors",
    "Sec-Fetch-Site": "cross-site",
    "Authorization": "",
    "Referer": "https://imob.valuegaia.com.br/",
    "Connection": "keep-alive",
    "TE": "trailers",
    "Pragma": "no-cache",
    "Cache-Control": "no-cache"
  }
}

http.get(req, (res) => {
  const { statusCode } = res;
  const contentType = res.headers['content-type'];

  let error;
  // Any 2xx status code signals a successful response but
  // here we're only checking for 200.
  if (statusCode !== 200) {
    error = new Error('Request Failed.\n' +
      `Status Code: ${statusCode}`);
  } else if (!/^application\/json/.test(contentType)) {
    error = new Error('Invalid content-type.\n' +
      `Expected application/json but received ${contentType}`);
  }
  if (error) {
    console.error(error.message);
    // Consume response data to free up memory
    res.resume();
    return;
  }

  res.setEncoding('utf8');
  let rawData = '';
  res.on('data', (chunk) => { rawData += chunk; });
  res.on('end', () => {
    try {
      const parsedData = JSON.parse(rawData);
      getPhotos(parsedData.hits[0])
    } catch (e) {
      console.error(e.message);
    }
  });
}).on('error', (e) => {
  console.error(`Got error: ${e.message}`);
});
console.log("DONE")

function getPhotos(resp: IngaiaResponse) {
  var dir = `../${resp.property_reference}`
  if (fs.existsSync(dir)) {
    return
  }
  fs.mkdirSync(dir);
  if (resp.photos.length > 0) {
    resp.photos.forEach((photo, index) => {
      https.request(photo.big, function (response) {
        var data = new stream();
        response.on('data', function (chunk) {
          data.push(chunk);
        });
        response.on('end', function () {
          fs.writeFileSync(`${dir}/${resp.property_reference}_${index}.jpg`, data.read());
        });
      }).end();
    });
  }
}
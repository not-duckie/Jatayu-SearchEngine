# Jatayu Search Engine

  

  

 Jatayu is an open source federated search engine that respects your privacy and takes no personal data from your machine. The search engine is implemented using <b>golang</b> and <b>elastic search</b>.

  

  

 The search engine further employs a custom written crawler that crawlers all the links present on the given webpage/website, that can later be used to search through.

  

 ![](<images/Pasted image 20211222095957.png>)

  

  

## Searching

  

The search page provides enhanced search results that are related to the searched query via phrase matching and Natural Language Processing. The matched results are structured to be displayed 10 at time and distributed with page numbers.

  
  

![](<images/Pasted image 20211222115510.png>)

  
  

The user can visit different page numbers by clicking on the page numbers at the bottom and relevant results will be shown.

  

![](<images/Pasted image 20211222120737.png>)

  

# Image Search

Jatayu also provides an image search feature that allows users to search through the images in the database and get relevant results.

  

Users can search images via file format such as jpg

  

![](<images/Pasted image 20211222121613.png>)

  

Further users can search through images via search query ex  `samsung`

![](<images/Pasted image 20211222121546.png>)

  
  
  
  

## Did You Mean

The `did you mean` feature comes in handy when the input data is close to the search result but not exactly the same. In that case Jatayu suggests the closed matching term and clicking the term will fetch the users the related results.

  

![](<images/Pasted image 20211222120526.png>)

  

## Suggestions

Jatayu uses Phrase Matching and NLP to suggest searches based on real time input from the user.

  

This works by using jquery to send input data to a backend api called `/autocomplete` to get top 10 suggestions.

  

![](<images/Pasted image 20211222121109.png>)

  
  

## Crawler

  

Jatayu contains a custom crawler that can crawl any website by visiting all the links present on the website. The crawler further categorizes the crawled data in text and images to further enhance the raw data.

  

In order to use the crawler just input any website into the input box and the crawler will start its work. The crawler work load is distributed thus you can use other features of Jatayu while the website is being crawled.

  

  

![](<images/Pasted image 20211222110750.png>)

  

  

## Why use Jatayu Over Others ?

  

  

Ever notice ads constantly following you around? That’s in part because Google tracks your searches and hides trackers on millions of websites. They make money off the private data stolen from you. Showing ads is just one way of exploiting that stolen private data.

  

  

By contrast, our private search engine doesn’t track your searches and  help you to keep your browsing history more private, as it should be. And that’s just the beginning — by using Jatayu you also escape the manipulation of the filter bubble and can use the Internet faster (after all that tracking code is disabled).

  

  

![](<images/Pasted image 20211222101240.png>)

  

  

# Getting Started

Downloading and getting started with Jatayu is very easy. All you need is working docker setup and a browser.
  

## Installing From Docker

1. Download the Docker Image off the Docker Hub

```bash
docker pull notduckie/jatayu-searchengine
```

2. Run the installed image with the following command

This is open a port 443 on your localhost. Visit the that  port with any browser and you will be able to access jatayu search engine.

```bash
docker run -p443:443 -it notduckie/jatayu-searchengine /opt/setup.sh
```

Visit URL := `https://127.0.0.1/`

#### Note
Since the server is running self signed ssl certificate so you will require to accept on the following page.

1. Click on `Advanced`
2. Click on `Proceed to 127.0.0.1`

![](<images/Pasted image 20211222152014.png>)

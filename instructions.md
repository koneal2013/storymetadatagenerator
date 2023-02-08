# Axios Back End Exercise

Axios’s ethos is Smart Brevity; we exist to make people smarter, faster. We value our readers’ time, so our
journalists pay attention to the word count of their stories. One feature on our site and app: we let our
readers know up front how long they’ll spend reading a story.

## How to submit your solution

When you have finished, please submit your solution back to the Github repository we have shared out with you. You do not need to fork the assignment repository or send us a copy. Please reach out to the recruiting coordinator to let them know once you have completed the final version of the exercise.

## Before You Start

We're not trying to get you to work for us for free, so please don't spend than more than 4 hours on this.
You can write a TODO doc that explains how you'd complete any tasks you don't get to.

## What You're Building

Axios exposes a public API which provides a JSON rendering of our news stories. Your challenge: write a tool to
provide the word count for each of the stories in the first N pages of our story stream, along with the average
adult reading time for each story.

Please deliver clear, readable, commented, runnable code that does the following:

* Takes one parameter: the number of pages of stream results to word count, defaulting to 1 but
  accepting any positive integer.
* Returns a valid JSON response which includes the following for each story in the requested pages of
  the story stream:
  * headline (from our API response)
  * permalink (from our API response)
  * word count (calculated by your code)
  * expected reading time, in minutes, rounded to the nearest whole number, cast to a string. (calculated by your code)
    * if a story's reading time would round to 0 minutes (and many of ours do!) please return the
      string "<1"
* The tool should have an execution time suitable for real-time use.

Please implement this in Golang (the language in which nearly all of our backend development currently takes place). Feel free to Google, use Stack Overflow, your favorite IDE, etc: this is how we work every day.

We recommend against overly-complex or Mechanical Turk-based solutions: the performance is almost certain
to be too slow for our purposes 😁.

## API Details

1. The URL for our story stream is at `https://api.axios.com/api/render/stream/content/`
    * Note that this response is paginated, returning 10 story ids per page.
1. The URL to request the JSON version of a single story is `https://api.axios.com/api/render/content/<id>`, where `<id>` is one of the ids from the story stream, e.g.
   `https://api.axios.com/api/render/content/aa030df2-0a82-4066-bfba-7aa2ef316b75/`
1. To calculate the word count, you'll need to walk through the `blocks` attribute in the story response from our API.
    * You’ll find the content of the story as a list of DraftJS “blocks".
    * The blocks are in [DraftJS content block format](https://draftjs.org/docs/api-reference-content-block) -- the `text` key in each block is really the only one you need worry about. It contains the plaintext of a single paragraph in the story.
    * **Do not directly use any of the wordcounts you receive from the API** (in the `wordcount`, `read_more_wordcount` or `before_read_more_worcount` keys) though you are welcome to compare them to your result to get a sense of whether you're in the right ballpark.
    * Same for the `keep_reading_mins` our API already returns -- don't just re-use that.
    * If you're curious, you can see how the story looks on the Axios site
      by visiting the link at the `permalink` key, e.g.
      `https://www.axios.com/2021/06/03/biden-ban-investment-china-surveillance-companies`

### Example Data Your Tool Might Return For One Story In The Stream

```text
{
  "e064aa87-ec46-44bf-9aa1-2b456d0501b1": {
    "word_count": 533,
    "reading_time": "2",
    "headline": "Holiday shopping season starts early amid inflation",
    "permalink": "https://www.axios.com/2022/08/30/walmart-toys-inflation-holiday-shopping-season"
  }
}
```

## Suggestions

* Don’t worry if your answer doesn’t exactly match our generated numbers, but if they don’t we’d
  love to hear your thoughts on why they differ.
* We love well-tested code. Plus, bonus points for structuring your solution in a way that it can
  be tested without an internet connection.
* Please note any assumptions you make.
* Please submit your solution back to your repository with instructions explaining how we can run it.

## Support

If you have any questions about the exercise, feel free to [email us](mailto:em-support@axios.com).

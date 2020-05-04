# RSS-Sync

RSS-Sync will sync you rss feed into support targets (Trello atm).
I use it as to sync my favourite podcasts and add them to my Trello board so I dont forget to listen.


* The input to the program is `feed.yaml` file that describes the rss, targets and the binding between them [FEED].
* It uses go template as templat engine together with [gomplate](https://docs.gomplate.ca/) to extend the functionality 
* open-integration pipeline - [read more](https://dev.to/olegsu/continuous-automation-with-open-integration-4f5a) about open-intergration 


For example:
```yaml
targets:
# Unique name of the target
- name: This Week List
    trello:
    # Trello API token - https://trello.com/app-key
    token: '{{ env.Getenv "TRELLO_TOKEN" }}'
    # Trello application ID - https://trello.com/app-key
    application-id: '{{ env.Getenv "TRELLO_APP_ID" }}'
    # Trello board id - get it from the URL
    board-id: '{{ env.Getenv "TRELLO_BOARD_ID" }}'
    # Trello list id - get if from https://trello.com/b/{board-id}.json
    list-id: '{{ env.Getenv "TRELLO_LIST_ID" }}'
    
    # Data about the card to be created
    card:
        title: 'Listen to: {{ .item.title }}'
        description: "Link: {{ .item.link }}\nDescription: {{ .item.description }}"
        # Lables ID's
        labels: []


rss:
# Unique name of the target
- name: Making History
  # RSS feed url 
  url: https://www.ranlevi.com/feed/mh_network_feed
  # set of filter to run on each RSS item
  # all the filter must to pass in order to pass the item to the target
  filter:
    
    # name of the filter can be anything
    # the value must be "true" at the end of the templating process in order to consider the filter as successful
    # only items that been released in the last 24 hours
    just-released: '{{ ((time.Now).Add (time.Hour -24)).Before (time.Parse "Mon, 02 Jan 2006 15:04:05 -0700" .item.published) }}'

#  In some cased the RSS feed is username-password protected  
#  auth:
#    username: '{{ env.Getenv "USERNAME" }}'
#    password: '{{ env.Getenv "PASSWORD" }}'    

# binding between rss and target 
bindings:
- name: Making History
  rss: Making History
  target: This Week List
```
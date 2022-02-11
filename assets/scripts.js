document.addEventListener('alpine:init', () => {
    Alpine.store('isLoading', false)

    Alpine.store('newEntityDialog', {
        opened: false,
        entity: {},

        targetLocation: null,

        init() {
            this.reset()
        },

        reset() {
            this.opened = false
            this.entity = {
                name: null,
                description: null,
                artifacts: null,
                location: null,
                metadata: {
                    quantity: null,
                    owners: null,
                    tags: null,
                },
            }
        },

        open(x) {
            this.targetLocation = x
            this.reset()
            this.opened = true
        },

        make() {
            this.entity.location = this.targetLocation
            this.entity.metadata.quantity = parseInt(this.entity.metadata.quantity)
            return this.entity
        }

    })

    Alpine.store('api', {
        newEntity(data) {
            // Creating a XHR object
            let xhr = new XMLHttpRequest();
            let url = "/api/entity";

            // open a connection
            xhr.open("POST", url, true);

            // Set the request header i.e. which type of content you are sending
            xhr.setRequestHeader("Content-Type", "application/json");

            // Create a state change callback
            xhr.onreadystatechange = function () {
                if (xhr.readyState === 4 && xhr.status == 200) {
                    return xhr.status
                }
            };

            // Converting JSON data to string
            var data = JSON.stringify(Alpine.store('newEntityDialog').make());

            // Sending data with the request
            xhr.send(data);
        }

    })

    Alpine.store('entities', {
        init() {
            this.currentEntity = 0
            this.load()
            this.loadLocationTree()
            Alpine.store('isLoading').this = false
        },

        currentEntity: 0,

        setCurrentEntity(x) {
            this.currentEntity = x
            this.load()
            this.loadLocationTree()
        },

        toggleViewAllLocation() {
            this.viewAllLocation = !this.viewAllLocation
            this.load()
        },

        viewAllLocation: false,

        entities: {},

        load() {
            let url = '/api/entity/find/children/' + this.currentEntity + '/full';

            if (this.viewAllLocation) {
                url = url + '/recursive'
            };

            readAll(url)
                .then(response => response.json())
                .then(response => { this.entities = response; });
        },

        locationtree: [],
        fullstate: {},

        readname(x) {
            if (x == 0) {
                return 'World'
            }
            return this.fullstate[x].name
        },

        loadLocationTree() {
            let url = '/api/entity';

            readAll(url)
                .then(response => response.json())
                .then(response => { this.fullstate = response; })
            this.locationtree = []
            this.recurseLocationTree(this.currentEntity)
            this.locationtree.reverse()
        },

        recurseLocationTree(x) {
            this.locationtree.push(x)
            if (x != 0) {
                elem = this.fullstate[x]
                this.recurseLocationTree(elem.location)
            }
        },

        hasChildren(x) {
            for (key in this.fullstate) {
                if (this.fullstate[key].location == x) {
                    return true
                }
            }
            return false
        },

        listChildLocations(x) {
            let childLocations = []
            for (key in this.fullstate) {
                if (this.fullstate[key].location == x) {
                    childLocations.push(this.fullstate[key])
                }
            }
            return childLocations
        }

    })

})

function dprint(x) {
    console.log(x)
}

function readAll(x) {
    return fetch(x)
}
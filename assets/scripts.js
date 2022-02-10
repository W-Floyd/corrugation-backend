document.addEventListener('alpine:init', () => {
    Alpine.store('isLoading', {
        on: true,

        toggle() {
            this.on = !this.on
        },

        makeFalse() {
            this.on = false
        },

        makeTrue() {
            this.on = true
        }

    })

    Alpine.store('entities', {
        init() {
            this.currentEntity = 0
            this.load()
            this.loadLocationTree()
            Alpine.store('isLoading').makeFalse()
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

            readJson(url)
                .then(response => response.json())
                .then(response => { this.entities = response; });
        },

        locationtree: [],
        fullstate: {},

        readname(x) {
            if (x == 0){
                return 'World'
            }
            return this.fullstate[x].Name
        },

        loadLocationTree() {
            let url = '/api/entity';

            readJson(url)
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
                this.recurseLocationTree(elem.Location)
            }
        },

        hasChildren(x) {
            for (key in this.fullstate) {
                if (this.fullstate[key].Location == x) {
                    return true
                }
            }
            return false
        },

        listChildLocations(x){
            let childLocations=[]
            for (key in this.fullstate) {
                if (this.fullstate[key].Location == x) {
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

function readJson(x) {
    return fetch(x)
}
param([Parameter(Mandatory=$true)][string]$FileBrowserRoot,[string]$StudioWeb='')
$ErrorActionPreference='Stop'
$FileBrowserRoot=(Resolve-Path -LiteralPath $FileBrowserRoot).Path
if(!(Test-Path -LiteralPath (Join-Path $FileBrowserRoot 'backend\internal\web\server.go'))){throw 'Expected the FileBrowser Drive source checkout'}
$adapterRoot=$PSScriptRoot
Copy-Item (Join-Path $adapterRoot 'studio.go') (Join-Path $FileBrowserRoot 'backend\internal\web\studio.go') -Force
Copy-Item (Join-Path $adapterRoot 'Studio.vue') (Join-Path $FileBrowserRoot 'frontend\src\views\Studio.vue') -Force
function Edit-Once([string]$Relative,[string]$Marker,[string]$Anchor,[string]$Replacement){
 $file=Join-Path $FileBrowserRoot $Relative
 $body=[IO.File]::ReadAllText($file).Replace("`r`n","`n")
 if($body.Contains($Marker)){return}
 if(!$body.Contains($Anchor)){throw "Cannot find insertion anchor in $Relative; inspect the current source version"}
 $index=$body.IndexOf($Anchor,[StringComparison]::Ordinal)
 [IO.File]::WriteAllText($file,$body.Substring(0,$index)+$Replacement+$body.Substring($index+$Anchor.Length))
}
Edit-Once 'backend\internal\web\httpRouter.go' 'withUser(studioFile)' 'api.HandleFunc("GET /resources", withUser(resourceGetHandler))' ('api.HandleFunc("GET /resources", withUser(resourceGetHandler))'+"`n"+'api.HandleFunc("GET /studio/file", withUser(studioFile))'+"`n"+'api.HandleFunc("PUT /studio/file", withUser(studioFile))')
Edit-Once 'backend\internal\web\server.go' 'registerStudioRoutes(router' 'configureHTTPRouter(router, api, publicRoutes, publicApi)' ('configureHTTPRouter(router, api, publicRoutes, publicApi)'+"`n"+'registerStudioRoutes(router,settings.Config.Http.BaseURL)')
Edit-Once 'frontend\src\router\index.ts' "name:'Studio'" 'const routes: RouteRecordRaw[] = [' ('const routes: RouteRecordRaw[] = ['+"`n"+"{path:'/studio',component:Layout,meta:{requiresAuth:true},children:[{path:'',name:'Studio',component:()=>import('@/views/Studio.vue')}]},")
Edit-Once 'frontend\src\store\getters.ts' "return 'studio'" 'if (pathname.startsWith(`/tools`)) return ''tools'';' ('if (pathname.startsWith(`/tools`)) return ''tools'';'+"`n    if (pathname === '/studio') return 'studio';")
Edit-Once 'frontend\src\views\bars\Default.vue' "return 'Studio'" 'getTopTitle() {' ("getTopTitle() {`n      if (state.route?.name === 'Studio') return 'Studio';")
# Sidebar availability; existing file opening can also use /studio?source=...&path=...
Edit-Once 'frontend\src\components\sidebar\General.vue' 'Image Studio' '</template>' ('<router-link v-if="!isShare &amp;&amp; isDataLoaded" to="/studio" style="display:flex;padding:14px 24px;color:inherit;text-decoration:none">Image Studio</router-link>'+"`n</template>")
Edit-Once 'frontend\src\components\ContextMenu.vue' 'canOpenStudio(){' '  computed: {' (@'
  computed: {
    canOpenStudio(){return !getters.isShare()&&this.selectedCount===1&&/\.(png|jpe?g|webp|gif|bmp|svg|json)$/i.test(this.firstSelected?.name||'');},
    canOpenStudioPreview(){return !getters.isShare()&&/\.(png|jpe?g|webp|gif|bmp|svg|json)$/i.test(state.req?.name||'');},
'@)
Edit-Once 'frontend\src\components\ContextMenu.vue' 'openStudio(){' '  methods: {' (@'
  methods: {
    openStudio(){const item=this.firstSelected;if(!item)return;mutations.closeHovers();this.$router.push({path:'/studio',query:{source:item.source||state.req.source,path:item.path}});},
    openStudioPreview(){mutations.closeHovers();this.$router.push({path:'/studio',query:{source:state.req.source,path:state.req.path}});},
'@)
$downloadAnchor='        <action'+"`n"+'          v-if="showDownload"'
Edit-Once 'frontend\src\components\ContextMenu.vue' 'v-if="canOpenStudio"' $downloadAnchor ('        <action v-if="canOpenStudio" icon="photo_filter" label="Edit in Studio" @action="openStudio" />'+"`n"+$downloadAnchor)
Edit-Once 'frontend\src\components\ContextMenu.vue' 'v-if="canOpenStudioPreview"' '<action icon="info" :label="$t(''general.info'')" @action="showInfoPrompt"/>' ('<action icon="info" :label="$t(''general.info'')" @action="showInfoPrompt"/>'+"`n"+'<action v-if="canOpenStudioPreview" icon="photo_filter" label="Edit in Studio" @action="openStudioPreview" />')
if($StudioWeb){$StudioWeb=(Resolve-Path -LiteralPath $StudioWeb).Path;if(!(Test-Path (Join-Path $StudioWeb 'index.html'))){throw 'StudioWeb must contain the built index.html'};$target=Join-Path $FileBrowserRoot 'backend\internal\web\embed\studio';New-Item -ItemType Directory -Path $target -Force|Out-Null;Copy-Item (Join-Path $StudioWeb '*') $target -Recurse -Force}
Write-Host 'Studio adapter installed. Rebuild FileBrowser frontend, copy generated frontend assets to embed, retain Studio/Office assets, then compile the Go backend.'

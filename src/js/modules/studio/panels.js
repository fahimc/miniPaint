export default class StudioPanels {
 toggle(selector){const panel=document.querySelector(selector);panel.hidden=!panel.hidden;}
 navigator(){this.toggle('.sidebar_right .preview');}
 color(){this.toggle('.sidebar_right .colors');}
 properties(){this.toggle('.sidebar_right .details');}
 layers(){this.toggle('.sidebar_right .layers');}
 reset(){document.querySelectorAll('.sidebar_right .block').forEach(panel=>panel.hidden=false);}
}

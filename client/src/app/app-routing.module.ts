import { NgModule } from '@angular/core';
import { RouterModule } from '@angular/router';

import { EntityComponent } from './entity/entity.component';
import { EntityListComponent } from './entity/entity-list.component';
import { EntityNewComponent } from './entity/entity-new.component';
import { TopComponent } from './top/top.component';

@NgModule({
  imports: [
    RouterModule.forRoot([
      { path: 'entity', component: EntityComponent,
        children: [
          { path: 'list', component: EntityListComponent, },
          { path: 'new', component: EntityNewComponent, },
          { path: '', redirectTo: 'list', pathMatch: 'full', },
        ],
      },
      { path: 'entity', component: EntityComponent },
      { path: '', component: TopComponent },
    ]),
  ],
  exports: [RouterModule],
})
export class AppRoutingModule { }

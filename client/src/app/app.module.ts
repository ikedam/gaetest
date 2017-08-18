import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';

import { NgbModule } from '@ng-bootstrap/ng-bootstrap';

import { AppComponent } from './app.component';
import { AppRoutingModule } from './app-routing.module';
import { EntityComponent } from './entity/entity.component';
import { EntityListComponent } from './entity/entity-list.component';
import { EntityNewComponent } from './entity/entity-new.component';
import { TopComponent } from './top/top.component';

@NgModule({
  declarations: [
    AppComponent,
    EntityComponent,
    EntityListComponent,
    EntityNewComponent,
    TopComponent,
  ],
  imports: [
    BrowserModule,
    NgbModule.forRoot(),
    ReactiveFormsModule,

    AppRoutingModule,
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
